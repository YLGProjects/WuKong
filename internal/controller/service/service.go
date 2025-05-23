/**
 *MIT License
 *
 *Copyright (c) 2025 ylgeeker
 *
 *Permission is hereby granted, free of charge, to any person obtaining a copy
 *of this software and associated documentation files (the "Software"), to deal
 *in the Software without restriction, including without limitation the rights
 *to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *copies of the Software, and to permit persons to whom the Software is
 *furnished to do so, subject to the following conditions:
 *
 *copies or substantial portions of the Software.
 *
 *THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *SOFTWARE.
**/

package service

import (
	"YLGProjects/WuKong/pkg/constant"
	"YLGProjects/WuKong/pkg/logger"
	"YLGProjects/WuKong/pkg/proto"

	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Service struct {
	proto.UnimplementedControllerServiceServer

	ctx         context.Context
	address     string
	serviceID   string
	connections map[string]*Connection
	mu          sync.RWMutex
}

func New(ctx context.Context, address string, serviceID string) *Service {
	return &Service{
		ctx:         ctx,
		address:     address,
		serviceID:   serviceID,
		connections: make(map[string]*Connection),
	}
}

func (s *Service) extractClientInfo(ctx context.Context) (string, map[string]string, error) {
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())

	metadata := make(map[string]string)

	// TODO: parse metadata from ctx
	_ = ctx

	return clientID, metadata, nil
}

func (s *Service) registerConnection(clientID string, conn *Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingConn, exists := s.connections[clientID]; exists {
		existingConn.close()
	}

	s.connections[clientID] = conn
}

func (s *Service) unregisterConnection(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, exists := s.connections[clientID]; exists {
		conn.close()
		delete(s.connections, clientID)
	}
}

func (s *Service) getConnection(clientID string) (*Connection, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, exists := s.connections[clientID]
	return conn, exists
}

func (s *Service) Connect(stream proto.ControllerService_ConnectServer) error {
	ctx := stream.Context()
	clientID, metadata, err := s.extractClientInfo(ctx)
	if err != nil {
		return err
	}

	conn := &Connection{
		ClientID:   clientID,
		Stream:     stream,
		SendChan:   make(chan *proto.AgentResponse, 100),
		LastActive: time.Now(),
		Metadata:   metadata,
	}

	s.registerConnection(clientID, conn)
	defer s.unregisterConnection(clientID)

	wg := &sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()
		conn.sendMessages(s.ctx)
	}()

	go func() {
		defer wg.Done()
		conn.receiveMessages(s.ctx)
	}()

	wg.Wait()
	return nil
}

func (s *Service) Run() error {

	kasp := keepalive.ServerParameters{
		Time:    constant.DefaultServerPingTime,
		Timeout: constant.DefaultPingTimeout,
	}

	kacp := keepalive.EnforcementPolicy{
		MinTime:             constant.DefaultKeepaliveMiniTime,
		PermitWithoutStream: true,
	}

	svr := grpc.NewServer(
		grpc.KeepaliveParams(kasp),
		grpc.KeepaliveEnforcementPolicy(kacp),
		grpc.MaxRecvMsgSize(constant.DefaultMaxReceiveMessageSize),
		grpc.MaxSendMsgSize(constant.DefaultMaxSendMessageSize),
	)

	proto.RegisterControllerServiceServer(svr, s)
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	logger.Info("Server listening at %v", lis.Addr())
	return svr.Serve(lis)
}
