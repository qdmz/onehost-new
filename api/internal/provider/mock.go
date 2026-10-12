package provider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
)

// Mock 是无外部依赖的模拟后端，便于在没有真实 LXD/Incus 节点时跑通全流程。
type Mock struct{}

var mockStore = struct {
	sync.Mutex
	items map[string]string // remoteID -> state
}{items: map[string]string{}}

func (m *Mock) Ping(ctx context.Context) error { return nil }

func (m *Mock) Create(ctx context.Context, spec Spec) (string, error) {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	id := "mock-" + hex.EncodeToString(b)
	mockStore.Lock()
	mockStore.items[id] = "running"
	mockStore.Unlock()
	return id, nil
}

func (m *Mock) set(id, st string) error {
	mockStore.Lock()
	defer mockStore.Unlock()
	if _, ok := mockStore.items[id]; !ok {
		return fmt.Errorf("实例不存在: %s", id)
	}
	mockStore.items[id] = st
	return nil
}

func (m *Mock) Start(ctx context.Context, id string) error   { return m.set(id, "running") }
func (m *Mock) Stop(ctx context.Context, id string) error    { return m.set(id, "stopped") }
func (m *Mock) Restart(ctx context.Context, id string) error { return m.set(id, "running") }

func (m *Mock) Delete(ctx context.Context, id string) error {
	mockStore.Lock()
	defer mockStore.Unlock()
	delete(mockStore.items, id)
	return nil
}

func (m *Mock) State(ctx context.Context, id string) (string, error) {
	mockStore.Lock()
	defer mockStore.Unlock()
	if st, ok := mockStore.items[id]; ok {
		return st, nil
	}
	return "", fmt.Errorf("实例不存在: %s", id)
}

func (m *Mock) IP(ctx context.Context, id string) (string, error) {
	if v := os.Getenv("ONEHOST_MOCK_IP"); v != "" {
		return v, nil
	}
	return "127.0.0.1", nil
}
