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
	"YLGProjects/WuKong/pkg/logger"
	"context"
	"math"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Registry struct {
	serviceId     string
	rootKey       string
	ttl           int64
	wg            sync.WaitGroup
	mu            sync.Mutex
	client        *clientv3.Client
	leaseId       clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

func NewRegistry(c *clientv3.Client, serviceId string, ttl int64) (*Registry, error) {
	return nil, nil
}

func (r *Registry) grant(ctx context.Context) error {

	leaseResp, err := r.client.Grant(ctx, r.ttl)
	if err != nil {
		return gerrors.New(gerrors.OperationFailure, err.Error())
	}
	r.leaseId = leaseResp.ID

	keepAliveChan, err := r.client.KeepAlive(ctx, r.leaseId)
	if err != nil {
		return gerrors.New(gerrors.OperationFailure, err.Error())
	}
	r.keepAliveChan = keepAliveChan

	r.wg.Add(2)
	go r.monitorKeepalive(ctx)
	go r.checkLeaseTTL(ctx)

	return nil
}

func (r *Registry) monitorKeepalive(ctx context.Context) error {

	defer r.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return nil

		case resp, ok := <-r.keepAliveChan:
			if !ok {
				r.wg.Add(1)
				go func(ctx context.Context) {
					defer r.wg.Done()
					r.recoverLease(ctx)
				}(ctx)

				return gerrors.New(gerrors.ComponentFailure, "service registry is offline")
			}

			logger.Debug("keepalive, ID:%v", resp.ID)
		}
	}

}

func (r *Registry) checkLeaseTTL(ctx context.Context) error {

	defer r.wg.Done()

	ttl := time.Duration(math.Floor(float64(r.ttl) / 2))
	ticker := time.NewTicker(ttl * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ttlResp, err := r.client.TimeToLive(ctx, r.leaseId)
		if err != nil || ttlResp.TTL <= 0 {
			r.wg.Add(1)
			go func(ctx context.Context) {
				defer r.wg.Done()
				r.recoverLease(ctx)
			}(ctx)

			return gerrors.New(gerrors.ComponentFailure, "service registry is offline")
		}
	}

	return nil
}

func (r *Registry) recoverLease(ctx context.Context) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	ttlResp, err := r.client.Lease.TimeToLive(ctx, r.leaseId)
	if err == nil && ttlResp.TTL > 0 {
		r.wg.Add(2)
		go r.monitorKeepalive(ctx)
		go r.checkLeaseTTL(ctx)
		return nil
	}

	return nil
}

func (r *Registry) RegisterService(ctx context.Context, serviceId, value string) error {

	serviceId = strings.TrimSpace(serviceId)
	value = strings.TrimSpace(value)

	if serviceId == "" {
		return gerrors.New(gerrors.InvalidParameter, "serviceId is required")
	}

	r.serviceId = serviceId

	_, err := r.client.Put(ctx, r.rootKey, value, clientv3.WithLease(r.leaseId))
	if err != nil {
		return gerrors.New(gerrors.OperationFailure, err.Error())
	}

	return nil
}

func (r *Registry) Set(ctx context.Context, key, value string) error {

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	if key == "" {
		return gerrors.New(gerrors.InvalidParameter, "key is required")
	}

	_, err := r.client.Put(ctx, key, value, clientv3.WithLease(r.leaseId))
	if err != nil {
		return gerrors.New(gerrors.OperationFailure, err.Error())
	}

	return nil
}

func (r *Registry) Close() {
	r.wg.Wait()
}
