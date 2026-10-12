package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"onehost/internal/model"
)

// LXD / Incus 统一 REST 客户端。
// 二者 API 高度兼容：均为 https://host:8443/1.0，客户端证书认证，异步操作返回 operation URL。
type lxdProvider struct {
	base   string
	client *http.Client
}

type lxdResp struct {
	Type      string          `json:"type"`
	Error     string          `json:"error"`
	ErrorCode int             `json:"error_code"`
	Operation string          `json:"operation"`
	Metadata  json.RawMessage `json:"metadata"`
}

func NewLXD(node *model.Node) (Provider, error) {
	base := strings.TrimRight(node.APIURL, "/")
	if base == "" {
		return nil, fmt.Errorf("节点 %s 未配置 API 地址", node.Name)
	}
	if !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	// 默认补 8443 端口
	if !strings.Contains(strings.TrimPrefix(base, "https://"), ":") {
		base = base + ":8443"
	}

	tlsCfg := &tls.Config{InsecureSkipVerify: true}
	if node.Cert != "" && node.Key != "" {
		cert, err := tls.X509KeyPair([]byte(node.Cert), []byte(node.Key))
		if err != nil {
			return nil, fmt.Errorf("客户端证书解析失败: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}
	// 若提供了服务端证书，则启用严格校验
	if node.ServerCert != "" {
		pool := x509.NewCertPool()
		if pool.AppendCertsFromPEM([]byte(node.ServerCert)) {
			tlsCfg.InsecureSkipVerify = false
			tlsCfg.RootCAs = pool
		}
	}

	return &lxdProvider{
		base: base,
		client: &http.Client{
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
			Timeout:   300 * time.Second,
		},
	}, nil
}

func (c *lxdProvider) do(ctx context.Context, method, path string, body interface{}) (*lxdResp, error) {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var lr lxdResp
	if err := json.Unmarshal(data, &lr); err != nil {
		return nil, fmt.Errorf("响应解析失败(HTTP %d): %s", resp.StatusCode, string(data))
	}
	if lr.Error != "" {
		return nil, fmt.Errorf("lxd 错误: %s", lr.Error)
	}
	if lr.Type == "async" && lr.Operation != "" {
		return c.waitOperation(ctx, lr.Operation)
	}
	return &lr, nil
}

func (c *lxdProvider) waitOperation(ctx context.Context, op string) (*lxdResp, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.base+op+"/wait", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var lr lxdResp
	if err := json.Unmarshal(data, &lr); err != nil {
		return nil, fmt.Errorf("操作响应解析失败: %s", string(data))
	}
	if lr.Type == "error" || lr.Error != "" {
		return nil, fmt.Errorf("操作失败: %s", lr.Error)
	}
	var meta struct {
		Status string `json:"status"`
		Err    string `json:"err"`
	}
	_ = json.Unmarshal(lr.Metadata, &meta)
	if meta.Status != "" && meta.Status != "Success" {
		return nil, fmt.Errorf("操作未成功(%s): %s", meta.Status, meta.Err)
	}
	return &lr, nil
}

func (c *lxdProvider) Ping(ctx context.Context) error {
	_, err := c.do(ctx, "GET", "/1.0", nil)
	return err
}

func (c *lxdProvider) Create(ctx context.Context, spec Spec) (string, error) {
	body := map[string]interface{}{
		"name": spec.Name,
		"type": "container",
		"source": map[string]interface{}{
			"type":     "image",
			"alias":    spec.Image,
			"protocol": "simplestreams",
			"server":   "https://images.linuxcontainers.org",
		},
		"config": map[string]string{
			"limits.cpu":    fmt.Sprintf("%d", spec.CPU),
			"limits.memory": fmt.Sprintf("%dMiB", spec.MemoryMB),
		},
		"devices": map[string]interface{}{
			"root": map[string]interface{}{
				"type": "disk",
				"path": "/",
				"pool": "default",
			},
		},
	}
	if _, err := c.do(ctx, "POST", "/1.0/instances", body); err != nil {
		return "", err
	}
	return spec.Name, nil
}

func (c *lxdProvider) stateAction(ctx context.Context, id, action string) error {
	body := map[string]string{"action": action, "timeout": "30"}
	_, err := c.do(ctx, "POST", "/1.0/instances/"+id+"/state", body)
	return err
}

func (c *lxdProvider) Start(ctx context.Context, id string) error   { return c.stateAction(ctx, id, "start") }
func (c *lxdProvider) Stop(ctx context.Context, id string) error    { return c.stateAction(ctx, id, "stop") }
func (c *lxdProvider) Restart(ctx context.Context, id string) error { return c.stateAction(ctx, id, "restart") }

func (c *lxdProvider) Delete(ctx context.Context, id string) error {
	_, err := c.do(ctx, "DELETE", "/1.0/instances/"+id, nil)
	return err
}

func (c *lxdProvider) rawState(ctx context.Context, id string) (json.RawMessage, error) {
	r, err := c.do(ctx, "GET", "/1.0/instances/"+id+"/state", nil)
	if err != nil {
		return nil, err
	}
	return r.Metadata, nil
}

func (c *lxdProvider) State(ctx context.Context, id string) (string, error) {
	meta, err := c.rawState(ctx, id)
	if err != nil {
		return "unknown", err
	}
	var m struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(meta, &m)
	return strings.ToLower(m.Status), nil
}

func (c *lxdProvider) IP(ctx context.Context, id string) (string, error) {
	meta, err := c.rawState(ctx, id)
	if err != nil {
		return "", err
	}
	var m struct {
		Network map[string]struct {
			Addresses []struct {
				Family  string `json:"family"`
				Address string `json:"address"`
				Scope   string `json:"scope"`
			} `json:"addresses"`
		} `json:"network"`
	}
	_ = json.Unmarshal(meta, &m)
	for _, iface := range m.Network {
		for _, a := range iface.Addresses {
			if a.Family == "inet" && a.Scope == "global" {
				return a.Address, nil
			}
		}
	}
	return "", fmt.Errorf("未获取到 IPv4 地址")
}
