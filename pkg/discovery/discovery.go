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
	"strings"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Discovery struct {
	exit   chan struct{}
	wg     sync.WaitGroup
	client *clientv3.Client
}

func (d *Discovery) Watch(ctx context.Context, key string) (chan *Event, error) {

	key = strings.TrimSpace(key)
	if key == "" {
		return nil, gerrors.New(gerrors.InvalidParameter, "the watched key is required")
	}

	watchChan := d.client.Watch(ctx, key)
	eventChan := make(chan *Event, 1024)

	d.wg.Add(1)

	go func(ctx context.Context) {

		defer d.wg.Done()

		for {
			select {
			case <-d.exit:
				logger.Info("exit the watcher. key:%s", key)
				return

			case <-ctx.Done():
				logger.Info("exit the watcher. key:%s", key)
				return

			case watchResp := <-watchChan:
				for _, event := range watchResp.Events {
					switch event.Type {
					case clientv3.EventTypePut:
						event := &Event{
							Type:  EventTypePut,
							Key:   string(event.Kv.Key),
							Value: []byte(event.Kv.Value),
						}
						eventChan <- event

					case clientv3.EventTypeDelete:
						event := &Event{
							Type:  EventTypeDelete,
							Key:   string(event.Kv.Key),
							Value: []byte(event.Kv.Value),
						}
						eventChan <- event
					}
				}
			}
		}
	}(ctx)

	return eventChan, nil
}

func (d *Discovery) WatchWithPrefix(ctx context.Context, keyPrefix string) (chan *Event, error) {

	keyPrefix = strings.TrimSpace(keyPrefix)
	if keyPrefix == "" {
		return nil, gerrors.New(gerrors.InvalidParameter, "the watched key is required")
	}

	watchChan := d.client.Watch(ctx, keyPrefix, clientv3.WithPrefix())
	eventChan := make(chan *Event, 1024)

	go func(ctx context.Context) {
		for {
			select {
			case <-d.exit:
				logger.Info("exit the watcher. key prefix:%s", keyPrefix)
				return

			case <-ctx.Done():
				logger.Info("exit the watcher. key prefix:%s", keyPrefix)
				return

			case watchResp := <-watchChan:
				for _, event := range watchResp.Events {
					switch event.Type {
					case clientv3.EventTypePut:
						event := &Event{
							Type:  EventTypePut,
							Key:   string(event.Kv.Key),
							Value: []byte(event.Kv.Value),
						}
						eventChan <- event

					case clientv3.EventTypeDelete:
						event := &Event{
							Type:  EventTypeDelete,
							Key:   string(event.Kv.Key),
							Value: []byte(event.Kv.Value),
						}
						eventChan <- event
					}
				}
			}
		}
	}(ctx)

	return eventChan, nil
}

func (d *Discovery) Get(ctx context.Context, key string) ([]byte, error) {

	key = strings.TrimSpace(key)
	if key == "" {
		return nil, gerrors.New(gerrors.InvalidParameter, "the key prefix is required")
	}

	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, gerrors.New(gerrors.ComponentFailure, err.Error())
	}

	var value []byte
	for _, kv := range resp.Kvs {
		value = []byte(kv.Value)
	}

	return value, nil
}

func (d *Discovery) GetWithPrefix(ctx context.Context, keyPrefix string) (map[string][]byte, error) {

	keyPrefix = strings.TrimSpace(keyPrefix)
	if keyPrefix == "" {
		return nil, gerrors.New(gerrors.InvalidParameter, "the key prefix is required")
	}

	resp, err := d.client.Get(ctx, keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, gerrors.New(gerrors.ComponentFailure, err.Error())
	}

	kvs := make(map[string][]byte, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		kvs[string(kv.Key)] = []byte(kv.Value)
	}

	return kvs, nil
}

func (d *Discovery) Close() {

	close(d.exit)
	d.wg.Wait()
}
