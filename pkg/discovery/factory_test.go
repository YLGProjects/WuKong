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

package discovery_test

import (
	"YLGProjects/WuKong/pkg/discovery"
	"context"
	"log"
	"os"
	"strings"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	client *clientv3.Client
	reg    *discovery.Registry
	dis    *discovery.Discovery
)

func setup() {

	endpoints := os.Getenv("YLG_ETCD_ENDPOINTS")
	user := os.Getenv("YLG_ETCDTL_USER")
	password := os.Getenv("YLG_ETCDTL_PASSWORLD")

	log.Println("endpoints:", endpoints)
	log.Println("user:", user)
	log.Println("password:", password)

	cli, err := discovery.NewClient(&discovery.ClientOptions{
		Endpoints: strings.Split(endpoints, ","),
		User:      user,
		Password:  password,
	})

	if err != nil {
		log.Fatalf("failed to create etcd client. errmsg:%v", err)
	}
	client = cli

	registry, err := discovery.NewRegistry(client, "test-service-id", 10)
	if err != nil {
		log.Fatalf("failed to create registry. errmsg:%v", err)
	}

	reg = registry

	discover, err := discovery.NewDiscovery(client)
	if err != nil {
		log.Fatalf("failed to create discovery. errmsg:%v", err)
	}

	dis = discover

	reg.SetService(context.Background(), "test-id")
}

func tear() {
	reg.Close()
	dis.Close()
	client.Close()
}

func TestSetService(t *testing.T) {

	err := reg.SetService(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("failed to set registry service. errmsg:%v", err)
	}
}

func TestGetWithPrefix(t *testing.T) {

	resp, err := dis.GetWithPrefix(context.Background(), "/")
	if err != nil {
		t.Fatalf("failed to get with prefix. errmsg:%v", err)
	}

	for key, value := range resp {
		t.Logf("key:%s, value:%s", key, string(value))
	}
}

func TestMain(m *testing.M) {

	setup()
	code := m.Run()
	tear()
	os.Exit(code)
}
