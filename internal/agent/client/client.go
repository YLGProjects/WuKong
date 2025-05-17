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

type Client struct {
	conn     *grpc.ClientConn
	client   proto.ConnectionServiceClient
	stream   proto.ConnectionService_ConnectClient
	ctx      context.Context
	cancel   context.CancelFunc
	clientID string
	mu       sync.RWMutex
	closed   bool

	messageHandler func(*proto.DataMessage)

	// 重连相关
	reconnecting         bool
	reconnectInterval    time.Duration
	maxReconnectAttempts int
	reconnectAttempts    int
}

func (c *Client) connect() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("client is closed")
	}
	c.mu.Unlock()

	stream, err := c.client.Connect(c.ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.stream = stream
	c.reconnectAttempts = 0
	c.mu.Unlock()

	go c.receiveMessages()

	go c.sendConnectionEstablished()

	go c.monitorConnection()

	return nil
}

func (c *Client) handleDisconnect() {
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

func (c *Client) monitorConnection() {
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

			state := c.GetConnectionState()
			c.mu.RUnlock()

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

func (c *Client) receiveMessages() {
	for {
		c.mu.RLock()
		if c.closed || c.stream == nil {
			c.mu.RUnlock()
			return
		}
		stream := c.stream
		c.mu.RUnlock()

		msg, err := stream.Recv()
		if err != nil {
			logger.Warn("Error receiving message: %v, starting reconnect", err)
			go c.handleDisconnect()
			return
		}

		if c.messageHandler != nil {
			c.messageHandler(msg)
		}

		logger.Debug("msg:%v", msg)
	}
}

func (c *Client) sendConnectionEstablished() {
	clientMsg := fmt.Sprintf("Client connected: %s", c.clientID)
	msg := &proto.DataMessage{
		Content: []byte(clientMsg),
		Headers: map[string]string{"connection_type": "persistent"},
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.stream == nil {
		return
	}

	if err := c.stream.Send(msg); err != nil {
		logger.Error("Error sending connection established message: %v", err)
	}
}

func (c *Client) SetMessageHandler(handler func(*proto.DataMessage)) {
	c.messageHandler = handler
}

func (c *Client) Run(ctx context.Context) error {
	err := c.connect()
	if err != nil {
		logger.Error("failed to connect remote server. errmsg:%v", err)
		return err
	}

outerLoop:
	for {
		select {
		case <-ctx.Done():
			break outerLoop
		}
	}

	return nil
}

func (c *Client) SendMessage(content string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	if c.stream == nil {
		return fmt.Errorf("stream not initialized")
	}

	msg := &proto.DataMessage{
		Content: []byte(content),
	}

	return c.stream.Send(msg)
}

func (c *Client) Close() {
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
func (c *Client) GetConnectionState() connectivity.State {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return connectivity.Shutdown
	}

	return c.conn.GetState()
}

func New(serverAddr string, clientID string) (*Client, error) {
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second, // ping the server every 10 seconds
		Timeout:             3 * time.Second,  // ping timeout
		PermitWithoutStream: true,             // enable send a ping command
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

	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		conn:     conn,
		client:   proto.NewConnectionServiceClient(conn),
		ctx:      ctx,
		cancel:   cancel,
		clientID: clientID,

		reconnectInterval:    5 * time.Second,
		maxReconnectAttempts: 10,
	}, nil
}
