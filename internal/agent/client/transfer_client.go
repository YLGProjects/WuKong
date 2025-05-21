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

package client

import (
	"YLGProjects/WuKong/pkg/constant"
	"YLGProjects/WuKong/pkg/logger"
	"YLGProjects/WuKong/pkg/proto"
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type TransferClient struct {
	conn                 *grpc.ClientConn
	client               proto.TransferServiceClient
	stream               proto.TransferService_PushDataClient
	ctx                  context.Context
	cancel               context.CancelFunc
	clientId             string
	mu                   sync.RWMutex
	closed               bool
	reconnecting         bool
	reconnectInterval    time.Duration
	maxReconnectAttempts int
	reconnectAttempts    int
}

func NewTransferClient(ctx context.Context, serverAddr string, clientId string) (*TransferClient, error) {
	kacp := keepalive.ClientParameters{
		Time:                constant.DefaultClientPingTime,
		Timeout:             constant.DefaultPingTimeout,
		PermitWithoutStream: true,
	}

	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024),
			grpc.MaxCallSendMsgSize(10*1024*1024),
		),
	)

	if err != nil {
		return nil, err
	}

	ctxCancel, cancel := context.WithCancel(ctx)

	return &TransferClient{
		conn:                 conn,
		client:               proto.NewTransferServiceClient(conn),
		ctx:                  ctxCancel,
		cancel:               cancel,
		clientId:             clientId,
		reconnectInterval:    5 * time.Second,
		maxReconnectAttempts: 10,
	}, nil
}

func (c *TransferClient) connect() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("client is closed")
	}
	c.mu.Unlock()

	stream, err := c.client.PushData(c.ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.stream = stream
	c.reconnectAttempts = 0
	c.mu.Unlock()

	go c.sendConnectionEstablished()

	go c.monitorConnection()

	return nil
}

func (c *TransferClient) handleDisconnect() {
	c.mu.Lock()
	if c.closed || c.reconnecting {
		c.mu.Unlock()
		return
	}

	c.reconnecting = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.reconnecting = false
		c.mu.Unlock()
	}()

	c.mu.Lock()
	c.reconnectAttempts++
	reconnectAttempts := c.reconnectAttempts
	maxAttempts := c.maxReconnectAttempts
	c.mu.Unlock()

	if maxAttempts > 0 && reconnectAttempts > maxAttempts {
		logger.Warn("Max reconnect attempts (%d) reached, giving up", maxAttempts)
		return
	}

	// The exponential backoff algorithm calculates the reconnection interval.
	backoffInterval := c.reconnectInterval * time.Duration(1<<uint(reconnectAttempts-1))

	// Add some randomness to avoid the stampede effect.
	backoffInterval = backoffInterval + time.Duration(rand.Int63n(int64(backoffInterval/2)))

	logger.Info("Reconnect attempt %d in %v", reconnectAttempts, backoffInterval)
	time.Sleep(backoffInterval)

	// retry
	logger.Info("Attempting to reconnect...")
	err := c.connect()
	if err != nil {
		logger.Warn("Reconnect failed: %v", err)
		go c.handleDisconnect()
	} else {
		logger.Info("Reconnect successful")
	}
}

func (c *TransferClient) monitorConnection() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.RLock()
			if c.closed {
				c.mu.RUnlock()
				return
			}
			c.mu.RUnlock()
			state := c.GetConnectionState()
			if state == connectivity.TransientFailure || state == connectivity.Shutdown {
				logger.Warn("Connection state: %s, starting reconnect", state.String())
				go c.handleDisconnect()
				return
			}

		case <-c.ctx.Done():
			return
		}
	}
}

func (c *TransferClient) sendConnectionEstablished() {
	msg := &proto.TransferRequest{}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.stream == nil {
		return
	}

	if err := c.stream.Send(msg); err != nil {
		logger.Error("Error sending connection established message: %v", err)
	}
}

func (c *TransferClient) Run() error {
	err := c.connect()
	if err != nil {
		logger.Error("failed to connect remote server. errmsg:%v", err)
		return err
	}

outerLoop:
	for {
		select {
		case <-c.ctx.Done():
			break outerLoop
		}
	}

	return nil
}

func (c *TransferClient) SendMessage(content string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	if c.stream == nil {
		return fmt.Errorf("stream not initialized")
	}

	msg := &proto.TransferRequest{}

	return c.stream.Send(msg)
}

func (c *TransferClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.closed = true

	if c.cancel != nil {
		c.cancel()
	}

	if c.conn != nil {
		c.conn.Close()
	}
}

// GetConnectionState Get the state of the connection.
func (c *TransferClient) GetConnectionState() connectivity.State {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return connectivity.Shutdown
	}

	return c.conn.GetState()
}
