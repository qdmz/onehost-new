// oh-sshd —— OneHost 联调用的真实 SSH 服务端。
// 完整实现 SSH 传输/认证/会话协商，支持：交互式 shell（真 PTY）、exec、SFTP 子系统。
// 用于在没有 LXD/Incus 的环境中，为 WebSSH / SFTP 提供端到端真实链路靶机。
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

type ptyReq struct {
	Term    string
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
	Modes   string
}
type winChange struct{ Columns, Rows, Width, Height uint32 }
type execReq struct{ Command string }
type subsysReq struct{ Name string }

type session struct {
	mu   sync.Mutex
	ptmx *os.File
	cols uint32
	rows uint32
}

func main() {
	addr := env("OH_SSH_ADDR", "0.0.0.0:2222")
	root := env("OH_SSH_ROOT", "/tmp/ohroot")
	user := env("OH_SSH_USER", "root")
	pass := env("OH_SSH_PASS", "onehost")

	if err := os.MkdirAll(root, 0o755); err != nil {
		log.Fatalf("创建根目录失败: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		log.Fatalf("切换根目录失败: %v", err)
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		log.Fatal(err)
	}

	cfg := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, p []byte) (*ssh.Permissions, error) {
			if c.User() == user && string(p) == pass {
				return nil, nil
			}
			return nil, fmt.Errorf("认证失败")
		},
	}
	cfg.AddHostKey(signer)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("oh-sshd 监听 %s  user=%s root=%s", addr, user, root)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func() {
			defer conn.Close()
			sconn, chans, reqs, err := ssh.NewServerConn(conn, cfg)
			if err != nil {
				return
			}
			log.Printf("SSH 登录: user=%s from=%s", sconn.User(), sconn.RemoteAddr())
			go ssh.DiscardRequests(reqs)
			for nc := range chans {
				if nc.ChannelType() != "session" {
					_ = nc.Reject(ssh.UnknownChannelType, "仅支持 session")
					continue
				}
				ch, chReqs, err := nc.Accept()
				if err != nil {
					continue
				}
				go handleSession(ch, chReqs)
			}
		}()
	}
}

func handleSession(ch ssh.Channel, reqs <-chan *ssh.Request) {
	defer ch.Close()
	st := &session{cols: 80, rows: 24}
	for req := range reqs {
		switch req.Type {
		case "pty-req":
			var p ptyReq
			if err := ssh.Unmarshal(req.Payload, &p); err == nil {
				st.mu.Lock()
				if p.Columns > 0 {
					st.cols = p.Columns
				}
				if p.Rows > 0 {
					st.rows = p.Rows
				}
				if st.ptmx != nil {
					_ = pty.Setsize(st.ptmx, &pty.Winsize{Cols: uint16(st.cols), Rows: uint16(st.rows)})
				}
				st.mu.Unlock()
			}
			_ = req.Reply(true, nil)
		case "window-change":
			var w winChange
			_ = ssh.Unmarshal(req.Payload, &w)
			st.mu.Lock()
			if w.Columns > 0 {
				st.cols = w.Columns
			}
			if w.Rows > 0 {
				st.rows = w.Rows
			}
			if st.ptmx != nil {
				_ = pty.Setsize(st.ptmx, &pty.Winsize{Cols: uint16(st.cols), Rows: uint16(st.rows)})
			}
			st.mu.Unlock()
		case "shell":
			_ = req.Reply(true, nil)
			go func() { runShell(ch, st); _ = ch.Close() }()
		case "exec":
			var e execReq
			_ = ssh.Unmarshal(req.Payload, &e)
			_ = req.Reply(true, nil)
			go func() { runExec(ch, e.Command); _ = ch.Close() }()
		case "subsystem":
			var s subsysReq
			_ = ssh.Unmarshal(req.Payload, &s)
			if s.Name == "sftp" {
				_ = req.Reply(true, nil)
				go func() { runSFTP(ch); _ = ch.Close() }()
			} else {
				_ = req.Reply(false, nil)
			}
		default:
			_ = req.Reply(false, nil)
		}
	}
}

func runShell(ch ssh.Channel, st *session) {
	cmd := exec.Command("/bin/bash", "-l")
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(st.cols), Rows: uint16(st.rows)})
	if err != nil {
		_, _ = ch.Write([]byte("无法启动 shell: " + err.Error() + "\r\n"))
		sendExit(ch, 1)
		return
	}
	defer func() {
		_ = ptmx.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_, _ = cmd.Process.Wait()
	}()
	st.mu.Lock()
	st.ptmx = ptmx
	st.mu.Unlock()

	go func() { _, _ = io.Copy(ptmx, ch) }()
	_, _ = io.Copy(ch, ptmx)
	sendExit(ch, 0)
}

func runExec(ch ssh.Channel, command string) {
	c := exec.Command("/bin/bash", "-c", command)
	c.Stdout = ch
	c.Stderr = ch.Stderr()
	status := uint32(0)
	if err := c.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			status = uint32(ee.ExitCode())
		} else {
			status = 1
		}
	}
	sendExit(ch, status)
}

func runSFTP(ch ssh.Channel) {
	srv, err := sftp.NewServer(ch)
	if err != nil {
		return
	}
	_ = srv.Serve()
	_ = srv.Close()
}

func sendExit(ch ssh.Channel, status uint32) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], status)
	_, _ = ch.SendRequest("exit-status", false, buf[:])
}
