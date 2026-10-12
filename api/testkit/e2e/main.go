// e2e —— OneHost 控制台端到端联调程序（真实协议，非模拟）。
// 覆盖：注册/登录 -> 充值 -> 下单 -> 支付开通 -> WebSSH(PTY) -> SFTP(列/传/下/删) -> VNC(RFB)。
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var (
	base   = env("E2E_API", "http://127.0.0.1:9999")
	wsBase = strings.Replace(strings.Replace(base, "https://", "wss://", 1), "http://", "ws://", 1)
	passed int
	failed int
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func check(name string, ok bool, detail string) {
	if ok {
		passed++
		fmt.Printf("  [PASS] %s\n", name)
	} else {
		failed++
		fmt.Printf("  [FAIL] %s -- %s\n", name, detail)
	}
}

func do(method, path, token string, body interface{}, out interface{}) error {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, base+path, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	var r envelope
	if err := json.Unmarshal(b, &r); err != nil {
		return fmt.Errorf("非法响应(HTTP %d): %s", res.StatusCode, string(b))
	}
	if r.Code != 0 {
		return fmt.Errorf("code=%d msg=%s", r.Code, r.Msg)
	}
	if out != nil && len(r.Data) > 0 {
		return json.Unmarshal(r.Data, out)
	}
	return nil
}

func main() {
	fmt.Println("== OneHost 控制台端到端联调 ==")
	fmt.Println("API:", base)

	// 1) 注册 / 登录
	email, pw := "e2e@onehost.local", "e2e-123456"
	_ = do("POST", "/api/v1/auth/register", "", map[string]string{"email": email, "password": pw}, nil)
	var login struct {
		Token string `json:"token"`
	}
	if err := do("POST", "/api/v1/auth/login", "", map[string]string{"email": email, "password": pw}, &login); err != nil {
		fmt.Println("登录失败，终止:", err)
		os.Exit(1)
	}
	token := login.Token
	check("登录获取 token", token != "", "空 token")

	// 2) 充值（模拟支付）
	var rc struct {
		Order struct {
			ID uint `json:"id"`
		} `json:"order"`
		Mock bool `json:"mock"`
	}
	if err := do("POST", "/api/v1/user/recharge", token, map[string]interface{}{"amount": 50000}, &rc); err != nil {
		fmt.Println("创建充值单失败:", err)
		os.Exit(1)
	}
	if err := do("POST", fmt.Sprintf("/api/v1/user/recharge/%d/mockpay", rc.Order.ID), token, nil, nil); err != nil {
		fmt.Println("模拟支付失败:", err)
		os.Exit(1)
	}
	var profile struct {
		Balance int64 `json:"balance"`
	}
	_ = do("GET", "/api/v1/user/profile", token, nil, &profile)
	check("充值到账余额 > 0", profile.Balance > 0, fmt.Sprintf("balance=%d", profile.Balance))

	// 3) 下单 -> 支付（触发开通）
	var products []struct {
		ID uint `json:"id"`
	}
	if err := do("GET", "/api/v1/products", "", nil, &products); err != nil || len(products) == 0 {
		fmt.Println("获取产品失败:", err)
		os.Exit(1)
	}
	var order struct {
		ID uint `json:"id"`
	}
	if err := do("POST", "/api/v1/user/orders", token, map[string]interface{}{"product_id": products[0].ID, "quantity": 1}, &order); err != nil {
		fmt.Println("下单失败:", err)
		os.Exit(1)
	}
	if err := do("POST", fmt.Sprintf("/api/v1/user/orders/%d/pay", order.ID), token, nil, nil); err != nil {
		fmt.Println("支付失败:", err)
		os.Exit(1)
	}
	check("下单并支付", true, "")

	// 4) 轮询实例就绪
	var instID uint
	var inst struct {
		ID      uint   `json:"id"`
		Status  string `json:"status"`
		IP      string `json:"ip"`
		SSHPort int    `json:"ssh_port"`
		SSHUser string `json:"ssh_user"`
		VNCPort int    `json:"vnc_port"`
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		var list []struct {
			ID      uint   `json:"id"`
			Status  string `json:"status"`
			IP      string `json:"ip"`
			SSHPort int    `json:"ssh_port"`
			SSHUser string `json:"ssh_user"`
			VNCPort int    `json:"vnc_port"`
		}
		if err := do("GET", "/api/v1/user/instances", token, nil, &list); err == nil {
			for _, it := range list {
				if it.Status == "running" {
					instID = it.ID
					inst.ID, inst.Status, inst.IP = it.ID, it.Status, it.IP
					inst.SSHPort, inst.SSHUser, inst.VNCPort = it.SSHPort, it.SSHUser, it.VNCPort
					break
				}
			}
		}
		if instID > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	check("实例开通且 running", instID > 0, "30s 内未就绪")
	if instID == 0 {
		os.Exit(1)
	}
	fmt.Printf("     实例 #%d ip=%s ssh=%d user=%s vnc=%d\n", inst.ID, inst.IP, inst.SSHPort, inst.SSHUser, inst.VNCPort)
	check("实例带控制台参数", inst.SSHPort > 0 && inst.VNCPort > 0, "缺失 ssh/vnc 参数")

	// 5) 控制台信息
	var cinfo map[string]interface{}
	check("控制台信息接口", do("GET", fmt.Sprintf("/api/v1/user/instances/%d/console", instID), token, nil, &cinfo) == nil, "调用失败")

	// 6) SFTP：列目录 -> 上传 -> 校验 -> 下载 -> 删除
	var listResp struct {
		Entries []struct {
			Name string `json:"name"`
		} `json:"entries"`
	}
	check("SFTP 列目录", do("GET", fmt.Sprintf("/api/v1/user/instances/%d/files?path=.", instID), token, nil, &listResp) == nil, "列目录失败")

	upOK := false
	{
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, _ := mw.CreateFormFile("file", "e2e-hello.txt")
		_, _ = fw.Write([]byte("hello-onehost-sftp"))
		_ = mw.Close()
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/user/instances/%d/files/upload?dir=.", base, instID), &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)
		if res, err := http.DefaultClient.Do(req); err == nil {
			b, _ := io.ReadAll(res.Body)
			var r envelope
			_ = json.Unmarshal(b, &r)
			upOK = r.Code == 0
			if !upOK {
				fmt.Println("     上传返回:", string(b))
			}
		}
	}
	check("SFTP 上传文件", upOK, "上传失败")

	downOK := false
	{
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/user/instances/%d/files/download?path=./e2e-hello.txt", base, instID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		if res, err := http.DefaultClient.Do(req); err == nil {
			b, _ := io.ReadAll(res.Body)
			downOK = string(b) == "hello-onehost-sftp"
			_ = res.Body.Close()
		}
	}
	check("SFTP 下载内容一致", downOK, "内容不匹配")

	delOK := false
	{
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/user/instances/%d/files?path=./e2e-hello.txt", base, instID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		if res, err := http.DefaultClient.Do(req); err == nil {
			b, _ := io.ReadAll(res.Body)
			var r envelope
			_ = json.Unmarshal(b, &r)
			delOK = r.Code == 0
			_ = res.Body.Close()
		}
	}
	check("SFTP 删除文件", delOK, "删除失败")

	// 7) WebSSH：真实 PTY 执行命令
	wsURL := fmt.Sprintf("%s/api/v1/user/instances/%d/ssh?token=%s&cols=120&rows=30", wsBase, instID, token)
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		check("WebSSH 建立连接", false, err.Error())
	} else {
		readyOK := false
		_ = c.SetReadDeadline(time.Now().Add(25 * time.Second))
		// 等待 ready
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			if mt == websocket.TextMessage && strings.Contains(string(msg), "ready") {
				readyOK = true
				break
			}
		}
		check("WebSSH 会话就绪", readyOK, "未收到 ready")

		_ = c.WriteMessage(websocket.BinaryMessage, []byte("echo OH_E2E_MARK_$((6*7))\n"))
		var acc strings.Builder
		cmdOK := false
		_ = c.SetReadDeadline(time.Now().Add(25 * time.Second))
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				break
			}
			if mt == websocket.BinaryMessage {
				acc.Write(msg)
				if strings.Contains(acc.String(), "OH_E2E_MARK_42") {
					cmdOK = true
					break
				}
			}
		}
		check("WebSSH 执行命令并回显", cmdOK, "未见预期输出")
		_ = c.WriteMessage(websocket.BinaryMessage, []byte("exit\n"))
		_ = c.Close()
	}

	// 8) VNC：RFB 握手
	vncURL := fmt.Sprintf("%s/api/v1/user/instances/%d/vnc?token=%s", wsBase, instID, token)
	vc, _, err := websocket.DefaultDialer.Dial(vncURL, nil)
	if err != nil {
		check("VNC 建立连接", false, err.Error())
	} else {
		_ = vc.SetReadDeadline(time.Now().Add(15 * time.Second))
		mt, msg, err := vc.ReadMessage()
		vncOK := err == nil && mt == websocket.BinaryMessage && strings.HasPrefix(string(msg), "RFB 003.008")
		check("VNC RFB 握手", vncOK, fmt.Sprintf("err=%v prefix=%.12q", err, string(msg)))
		_ = vc.Close()
	}

	fmt.Printf("\n结果: %d 通过, %d 失败\n", passed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
