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
	"context"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Discovery struct {
	client *clientv3.Client
}

func NewDiscovery(c *clientv3.Client) (*Discovery, error) {
	return nil, nil
}

func (d *Discovery) Watch(ctx context.Context, key string) error {

	key = strings.TrimSpace(key)
	if key == "" {
		return gerrors.New(gerrors.InvalidParameter, "the watched key is required")
	}

	watchChan := d.client.Watch(ctx, key)

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case clientv3.EventTypePut:
			case clientv3.EventTypeDelete:
			default:
			}
		}
	}

	return nil
}

func (d *Discovery) WatchWithPrefix(ctx context.Context, keyPrefix string) error {

	keyPrefix = strings.TrimSpace(keyPrefix)
	if keyPrefix == "" {
		return gerrors.New(gerrors.InvalidParameter, "the watched key is required")
	}

	watchChan := d.client.Watch(ctx, keyPrefix, clientv3.WithPrefix())

	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case clientv3.EventTypePut:
			case clientv3.EventTypeDelete:
			default:
			}
		}
	}

	return nil
}

func (d *Discovery) Get(ctx context.Context, key string) (string, error) {

	key = strings.TrimSpace(key)
	if key == "" {
		return "", gerrors.New(gerrors.InvalidParameter, "the key prefix is required")
	}

	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return "", gerrors.New(gerrors.ComponentFailure, err.Error())
	}

	value := ""
	for _, kv := range resp.Kvs {
		value = string(kv.Value)
	}

	return value, nil
}

func (d *Discovery) GetWithPrefix(ctx context.Context, keyPrefix string) (map[string]string, error) {

	keyPrefix = strings.TrimSpace(keyPrefix)
	if keyPrefix == "" {
		return nil, gerrors.New(gerrors.InvalidParameter, "the key prefix is required")
	}

	resp, err := d.client.Get(ctx, keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, gerrors.New(gerrors.ComponentFailure, err.Error())
	}

	kvs := make(map[string]string, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		kvs[string(kv.Key)] = string(kv.Value)
	}

	return kvs, nil
}
