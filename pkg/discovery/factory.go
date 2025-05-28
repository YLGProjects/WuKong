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

package discovery

import (
	"YLGProjects/WuKong/pkg/gerrors"
	"fmt"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EventType int

const (
	EventTypePut EventType = iota
	EventTypeDelete
	EventTypeRecover
)

const (
	defaultTTL                   = 6
	defaultDialTimeout           = 5 * time.Second
	defaultAutoSyncInterval      = 60 * time.Second
	defaultKeepAliveTime         = 30 * time.Second
	defaultKeepAliveTimeout      = 10 * time.Second
	defautlRegistryRootKeyPrefix = "/ylg/wukong/registry"
)

type Event struct {
	Type  EventType
	Key   string
	Value []byte
}
type ClientOptions struct {
	Endpoints            []string
	User                 string
	Password             string
	DialTimeout          time.Duration
	AutoSyncInterval     time.Duration
	DialKeepAliveTime    time.Duration
	DialKeepAliveTimeout time.Duration
}

func NewClient(opt *ClientOptions) (*clientv3.Client, error) {

	if opt == nil {
		return nil, gerrors.New(gerrors.InvalidParameter, "")
	}

	if len(opt.Endpoints) == 0 {
		return nil, gerrors.New(gerrors.InvalidParameter, "endpoints are required")
	}

	if opt.User == "" {
		return nil, gerrors.New(gerrors.InvalidParameter, "user is required")
	}

	if opt.Password == "" {
		return nil, gerrors.New(gerrors.InvalidParameter, "password is required")
	}

	if opt.DialTimeout == 0 {
		opt.DialTimeout = defaultDialTimeout
	}

	if opt.AutoSyncInterval == 0 {
		opt.AutoSyncInterval = defaultAutoSyncInterval
	}

	if opt.DialKeepAliveTime == 0 {
		opt.DialKeepAliveTime = defaultKeepAliveTime
	}

	if opt.DialKeepAliveTimeout == 0 {
		opt.DialKeepAliveTimeout = defaultKeepAliveTimeout
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            opt.Endpoints,
		DialTimeout:          opt.DialTimeout,
		AutoSyncInterval:     opt.AutoSyncInterval,
		DialKeepAliveTime:    opt.DialKeepAliveTime,
		DialKeepAliveTimeout: opt.DialKeepAliveTimeout,
	})

	if err != nil {
		return nil, gerrors.New(gerrors.ComponentFailure, err.Error())
	}

	return cli, nil
}

func NewRegistry(c *clientv3.Client, serviceId string, ttl int64) (*Registry, error) {

	if c == nil {
		return nil, gerrors.New(gerrors.InvalidParameter, "client is required")
	}

	serviceId = strings.TrimSpace(serviceId)
	if serviceId == "" {
		return nil, gerrors.New(gerrors.InvalidParameter, "service id is required")
	}

	if ttl < defaultTTL {
		ttl = defaultTTL
	}

	registry := &Registry{
		serviceId: serviceId,
		eventChan: make(chan *Event, 1024),
		rootKey:   fmt.Sprintf("%s/%s", defautlRegistryRootKeyPrefix, serviceId),
		ttl:       ttl,
		client:    c,
	}

	return registry, nil
}

func NewDiscovery(c *clientv3.Client) (*Discovery, error) {
	return nil, nil
}
