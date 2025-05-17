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
	"YLGProjects/WuKong/pkg/logger"
	"YLGProjects/WuKong/pkg/proto"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Service struct {
	proto.UnimplementedConnectionServiceServer

	address     string
	serviceID   string
	connections map[string]*Connection
	mu          sync.RWMutex
}

func (cm *Service) registerConnection(clientID string, conn *Connection) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if existingConn, exists := cm.connections[clientID]; exists {
		existingConn.close()
	}

	cm.connections[clientID] = conn
}

func (cm *Service) unregisterConnection(clientID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.connections[clientID]; exists {
		conn.close()
		delete(cm.connections, clientID)
	}
}

func (cm *Service) getConnection(clientID string) (*Connection, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[clientID]
	return conn, exists
}

func extractClientInfo(ctx context.Context) (string, map[string]string, error) {
	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())

	metadata := make(map[string]string)

	// TODO: parse metadata from ctx
	_ = ctx

	return clientID, metadata, nil
}

func (cm *Service) Connect(stream proto.ConnectionService_ConnectServer) error {
	ctx := stream.Context()
	clientID, metadata, err := extractClientInfo(ctx)
	if err != nil {
		return err
	}

	conn := &Connection{
		ClientID:   clientID,
		Stream:     stream,
		SendChan:   make(chan *proto.DataMessage, 100),
		LastActive: time.Now(),
		Metadata:   metadata,
	}

	cm.registerConnection(clientID, conn)
	defer cm.unregisterConnection(clientID)

	wg := &sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()
		conn.sendMessages()
	}()

	go func() {
		defer wg.Done()
		conn.receiveMessages()
	}()

	wg.Wait()
	return nil
}

// PushMessage push a message to the client
func (cm *Service) PushMessage(ctx context.Context, req *proto.PushRequest) (*proto.PushResponse, error) {
	clientID := req.GetClientID()

	conn, exists := cm.getConnection(clientID)
	if !exists {
		return &proto.PushResponse{Status: 404, Message: "Client not found"}, nil
	}

	select {
	case conn.SendChan <- req.GetMessage():
		return &proto.PushResponse{Status: 200, Message: "Message sent"}, nil
	case <-time.After(2 * time.Second):
		return &proto.PushResponse{Status: 500, Message: "Send timeout"}, nil
	}
}

func (s *Service) Run() error {

	kasp := keepalive.ServerParameters{
		MaxConnectionIdle:     30 * time.Minute,
		MaxConnectionAge:      2 * time.Hour,
		MaxConnectionAgeGrace: 5 * time.Minute,
		Time:                  5 * time.Minute,
		Timeout:               10 * time.Second,
	}

	kacp := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}

	svr := grpc.NewServer(
		grpc.KeepaliveParams(kasp),
		grpc.KeepaliveEnforcementPolicy(kacp),
		grpc.MaxConcurrentStreams(100),    // 每个连接的最大并发流数
		grpc.MaxRecvMsgSize(1024*1024*10), // 最大接收消息大小10MB
		grpc.MaxSendMsgSize(1024*1024*10), // 最大发送消息大小10MB
	)

	proto.RegisterConnectionServiceServer(svr, s)
	reflection.Register(svr)
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	logger.Info("Server listening at %v", lis.Addr())
	return svr.Serve(lis)
}

func New(address string, serviceID string) *Service {
	return &Service{
		address:     address,
		serviceID:   serviceID,
		connections: make(map[string]*Connection),
	}
}
