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

package constant

import "time"

const (
	DefaultClientPingTime             = 5 * time.Second
	DefaultServerPingTime             = 5 * time.Minute
	DefaultPingTimeout                = 10 * time.Second
	DefaultKeepaliveMiniTime          = 5 * time.Minute
	DefaultMaxReceiveMessageSize      = 1024 * 1024 * 10
	DefaultMaxSendMessageSize         = 1024 * 1024 * 10
	DefaultClientReconnectInterval    = 5 * time.Second
	DefaultClientMaxReconnectAttempts = 10
)
