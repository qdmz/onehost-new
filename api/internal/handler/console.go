package handler

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"onehost/internal/model"
	"onehost/internal/util"
)

// 实例控制台：WebSSH（SSH + PTY）、SFTP 文件管理、VNC（RFB over WebSocket）。
// 全部走真实协议，非模拟。

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// dialSSH 依据实例保存的控制台信息建立真实 SSH 连接。
func (h *Handler) dialSSH(inst *model.Instance) (*ssh.Client, error) {
	host := inst.IP
	if host == "" {
		host = "127.0.0.1"
	}
	port := inst.SSHPort
	if port <= 0 {
		port = 22
	}
	user := inst.SSHUser
	if user == "" {
		user = "root"
	}
	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(inst.SSHPassword)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // MVP：内网自建节点，跳过主机密钥校验
		Timeout:         15 * time.Second,
	}
	return ssh.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)), cfg)
}

func (h *Handler) sftpFor(c *gin.Context) (*ssh.Client, *sftp.Client, bool) {
	inst, ok := h.instanceForUser(c)
	if !ok {
		return nil, nil, false
	}
	if inst.Status != "running" {
		util.Fail(c, "实例未运行")
		return nil, nil, false
	}
	cli, err := h.dialSSH(inst)
	if err != nil {
		util.Fail(c, "SSH 连接失败: "+err.Error())
		return nil, nil, false
	}
	sc, err := sftp.NewClient(cli)
	if err != nil {
		_ = cli.Close()
		util.Fail(c, "SFTP 初始化失败: "+err.Error())
		return nil, nil, false
	}
	return cli, sc, true
}

// InstanceConsoleInfo 返回实例控制台接入信息（供前端决定入口可用性）。
func (h *Handler) InstanceConsoleInfo(c *gin.Context) {
	inst, ok := h.instanceForUser(c)
	if !ok {
		return
	}
	ready := inst.Status == "running" && inst.IP != ""
	user := inst.SSHUser
	if user == "" {
		user = "root"
	}
	util.OK(c, gin.H{
		"status": inst.Status,
		"ssh":    gin.H{"host": inst.IP, "port": inst.SSHPort, "user": user, "ready": ready},
		"vnc":    gin.H{"host": inst.IP, "port": inst.VNCPort, "ready": ready && inst.VNCPort > 0},
	})
}

func wsWriteJSON(ws *websocket.Conn, mu *sync.Mutex, v interface{}) {
	b, _ := json.Marshal(v)
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}
	_ = ws.WriteMessage(websocket.TextMessage, b)
}

// InstanceSSH WebSSH：浏览器 <-> SSH 会话（含 PTY）双向桥接。
// 文本帧=控制指令(JSON)，二进制帧=终端原始字节流。
func (h *Handler) InstanceSSH(c *gin.Context) {
	inst, ok := h.instanceForUser(c)
	if !ok {
		return
	}
	if inst.Status != "running" {
		util.Fail(c, "实例未运行，无法连接控制台")
		return
	}
	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "120"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "30"))
	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 30
	}

	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	var wmu sync.Mutex
	fail := func(msg string) { wsWriteJSON(ws, &wmu, gin.H{"type": "error", "msg": msg}) }

	cli, err := h.dialSSH(inst)
	if err != nil {
		fail("SSH 连接失败: " + err.Error())
		return
	}
	defer cli.Close()

	sess, err := cli.NewSession()
	if err != nil {
		fail("创建会话失败: " + err.Error())
		return
	}
	defer sess.Close()

	modes := ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}
	if err := sess.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		fail("申请终端失败: " + err.Error())
		return
	}
	stdin, err := sess.StdinPipe()
	if err != nil {
		fail("打开输入通道失败: " + err.Error())
		return
	}
	stdout, err := sess.StdoutPipe()
	if err != nil {
		fail("打开输出通道失败: " + err.Error())
		return
	}
	stderr, err := sess.StderrPipe()
	if err != nil {
		fail("打开错误通道失败: " + err.Error())
		return
	}
	if err := sess.Shell(); err != nil {
		fail("启动 shell 失败: " + err.Error())
		return
	}

	wsWriteJSON(ws, &wmu, gin.H{"type": "ready"})

	done := make(chan struct{})
	var once sync.Once
	finish := func() { once.Do(func() { close(done) }) }

	writeBin := func(p []byte) {
		wmu.Lock()
		_ = ws.WriteMessage(websocket.BinaryMessage, p)
		wmu.Unlock()
	}

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				writeBin(buf[:n])
			}
			if err != nil {
				finish()
				return
			}
		}
	}()
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				writeBin(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()

	go func() {
		for {
			mt, p, err := ws.ReadMessage()
			if err != nil {
				finish()
				return
			}
			switch mt {
			case websocket.BinaryMessage:
				_, _ = stdin.Write(p)
			case websocket.TextMessage:
				var m struct {
					Type string `json:"type"`
					Cols int    `json:"cols"`
					Rows int    `json:"rows"`
				}
				if json.Unmarshal(p, &m) == nil && m.Type == "resize" && m.Cols > 0 && m.Rows > 0 {
					_ = sess.WindowChange(m.Rows, m.Cols)
				}
			}
		}
	}()

	go func() {
		err := sess.Wait()
		code := 0
		if err != nil {
			if ee, ok := err.(*ssh.ExitError); ok {
				code = ee.ExitStatus()
			}
		}
		wsWriteJSON(ws, &wmu, gin.H{"type": "exit", "code": code})
		finish()
	}()

	<-done
}

// InstanceVNC：浏览器 <-> 实例 VNC(RFB) 的原始字节双向代理（兼容 noVNC）。
func (h *Handler) InstanceVNC(c *gin.Context) {
	inst, ok := h.instanceForUser(c)
	if !ok {
		return
	}
	if inst.Status != "running" || inst.VNCPort <= 0 {
		util.Fail(c, "实例未启用 VNC 控制台")
		return
	}
	ws, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	target := net.JoinHostPort(inst.IP, strconv.Itoa(inst.VNCPort))
	tcp, err := net.DialTimeout("tcp", target, 10*time.Second)
	if err != nil {
		wsWriteJSON(ws, nil, gin.H{"type": "error", "msg": "VNC 连接失败: " + err.Error()})
		return
	}
	defer tcp.Close()

	done := make(chan struct{})
	var once sync.Once
	finish := func() { once.Do(func() { close(done) }) }

	go func() {
		buf := make([]byte, 16384)
		for {
			n, err := tcp.Read(buf)
			if n > 0 {
				_ = ws.WriteMessage(websocket.BinaryMessage, buf[:n])
			}
			if err != nil {
				finish()
				return
			}
		}
	}()
	go func() {
		for {
			mt, p, err := ws.ReadMessage()
			if err != nil {
				finish()
				return
			}
			if mt == websocket.BinaryMessage {
				_, _ = tcp.Write(p)
			}
		}
	}()

	<-done
}

// ---------------- SFTP 文件管理 ----------------

type fileEntry struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	IsDir   bool   `json:"is_dir"`
	ModTime string `json:"mod_time"`
}

func (h *Handler) SFTPList(c *gin.Context) {
	cli, sc, ok := h.sftpFor(c)
	if !ok {
		return
	}
	defer cli.Close()
	defer sc.Close()

	p := c.DefaultQuery("path", ".")
	entries, err := sc.ReadDir(p)
	if err != nil {
		util.Fail(c, "读取目录失败: "+err.Error())
		return
	}
	out := make([]fileEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, fileEntry{
			Name:    e.Name(),
			Size:    e.Size(),
			Mode:    e.Mode().String(),
			IsDir:   e.IsDir(),
			ModTime: e.ModTime().Format(time.RFC3339),
		})
	}
	wd, _ := sc.Getwd()
	util.OK(c, gin.H{"path": p, "cwd": wd, "entries": out})
}

func (h *Handler) SFTPDownload(c *gin.Context) {
	cli, sc, ok := h.sftpFor(c)
	if !ok {
		return
	}
	defer cli.Close()
	defer sc.Close()

	p := c.Query("path")
	if p == "" {
		util.Fail(c, "缺少 path")
		return
	}
	f, err := sc.Open(p)
	if err != nil {
		util.Fail(c, "打开文件失败: "+err.Error())
		return
	}
	defer f.Close()
	st, _ := f.Stat()
	if st != nil && st.IsDir() {
		util.Fail(c, "目标是目录，无法下载")
		return
	}
	name := path.Base(p)
	c.Header("Content-Disposition", "attachment; filename=\""+name+"\"")
	c.Header("Content-Type", "application/octet-stream")
	if st != nil {
		c.Header("Content-Length", strconv.FormatInt(st.Size(), 10))
	}
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, f)
}

func (h *Handler) SFTPUpload(c *gin.Context) {
	cli, sc, ok := h.sftpFor(c)
	if !ok {
		return
	}
	defer cli.Close()
	defer sc.Close()

	dir := c.DefaultQuery("dir", ".")
	fh, err := c.FormFile("file")
	if err != nil {
		util.Fail(c, "缺少上传文件")
		return
	}
	src, err := fh.Open()
	if err != nil {
		util.Fail(c, "读取上传内容失败")
		return
	}
	defer src.Close()

	dst, err := sc.Create(path.Join(dir, path.Base(fh.Filename)))
	if err != nil {
		util.Fail(c, "创建远程文件失败: "+err.Error())
		return
	}
	defer dst.Close()
	n, err := io.Copy(dst, src)
	if err != nil {
		util.Fail(c, "写入失败: "+err.Error())
		return
	}
	util.OK(c, gin.H{"name": fh.Filename, "size": n})
}

func (h *Handler) SFTPMkdir(c *gin.Context) {
	cli, sc, ok := h.sftpFor(c)
	if !ok {
		return
	}
	defer cli.Close()
	defer sc.Close()

	p := c.Query("path")
	if p == "" {
		util.Fail(c, "缺少 path")
		return
	}
	if err := sc.Mkdir(p); err != nil {
		util.Fail(c, "创建目录失败: "+err.Error())
		return
	}
	util.OK(c, nil)
}

func (h *Handler) SFTPDelete(c *gin.Context) {
	cli, sc, ok := h.sftpFor(c)
	if !ok {
		return
	}
	defer cli.Close()
	defer sc.Close()

	p := c.Query("path")
	if p == "" {
		util.Fail(c, "缺少 path")
		return
	}
	st, err := sc.Stat(p)
	if err != nil {
		util.Fail(c, "对象不存在: "+err.Error())
		return
	}
	if st.IsDir() {
		if err := sc.RemoveDirectory(p); err != nil {
			util.Fail(c, "删除目录失败: "+err.Error())
			return
		}
	} else {
		if err := sc.Remove(p); err != nil {
			util.Fail(c, "删除文件失败: "+err.Error())
			return
		}
	}
	util.OK(c, nil)
}
