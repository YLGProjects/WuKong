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
	"sync"
	"time"
)

type Connection struct {
	ClientID   string
	Stream     proto.ConnectionService_ConnectServer
	SendChan   chan *proto.DataMessage
	LastActive time.Time
	Metadata   map[string]string
	mu         sync.RWMutex
	closed     bool
}

func (cc *Connection) sendMessages() {
	for msg := range cc.SendChan {
		if cc.isClosed() {
			return
		}

		cc.updateLastActive()

		if err := cc.Stream.Send(msg); err != nil {
			cc.close()
			return
		}
	}
}

func (cc *Connection) receiveMessages() {
	for {
		if cc.isClosed() {
			return
		}

		msg, err := cc.Stream.Recv()
		if err != nil {
			cc.close()
			return
		}

		cc.updateLastActive()

		logger.Debug("Received from %s: %s\n", cc.ClientID, msg.GetContent())
	}
}

func (cc *Connection) updateLastActive() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.LastActive = time.Now()
}

func (cc *Connection) isClosed() bool {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.closed
}

func (cc *Connection) close() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if !cc.closed {
		cc.closed = true
		close(cc.SendChan)
	}
}
