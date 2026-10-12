package provider

import (
	"context"
	"fmt"

	"onehost/internal/model"
)

// Spec 实例规格。
type Spec struct {
	Name     string
	Image    string // 镜像别名，如 images:debian/12
	CPU      int
	MemoryMB int
	DiskGB   int
}

// Provider 统一的虚拟化后端接口。
type Provider interface {
	Ping(ctx context.Context) error
	Create(ctx context.Context, spec Spec) (remoteID string, err error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Restart(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	State(ctx context.Context, id string) (string, error) // running / stopped / unknown
	IP(ctx context.Context, id string) (string, error)
}

// ForNode 根据节点类型返回对应实现。
func ForNode(node *model.Node) (Provider, error) {
	switch node.Type {
	case "lxd", "incus":
		return NewLXD(node)
	case "mock", "":
		return &Mock{}, nil
	default:
		return nil, fmt.Errorf("暂不支持的节点类型: %s", node.Type)
	}
}
