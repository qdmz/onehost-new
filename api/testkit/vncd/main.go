// oh-vncd —— OneHost 联调用的最小 RFB (VNC) 3.8 服务端。
// 完成版本/安全/初始化握手，并对 FramebufferUpdateRequest 返回一张渐变帧缓冲（raw 编码），
// 用于在没有真实 VNC 组件的环境中验证 VNC 通道端到端可用（兼容 noVNC）。
package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func main() {
	addr := env("OH_VNC_ADDR", "0.0.0.0:5901")
	const w, h = 640, 480

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("oh-vncd 监听 %s (%dx%d)", addr, w, h)
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go serve(c, w, h)
	}
}

func serve(c net.Conn, w, h int) {
	defer c.Close()
	br := bufio.NewReader(c)

	// 1) 服务端版本
	if _, err := c.Write([]byte("RFB 003.008\n")); err != nil {
		return
	}
	ver := make([]byte, 12)
	if _, err := io.ReadFull(br, ver); err != nil {
		return
	}
	// 3) 安全类型：仅 None(1)
	if _, err := c.Write([]byte{1, 1}); err != nil {
		return
	}
	choice := make([]byte, 1)
	if _, err := io.ReadFull(br, choice); err != nil {
		return
	}
	// 5) SecurityResult = OK
	if _, err := c.Write([]byte{0, 0, 0, 0}); err != nil {
		return
	}
	// 6) ClientInit
	ci := make([]byte, 1)
	if _, err := io.ReadFull(br, ci); err != nil {
		return
	}
	// 7) ServerInit
	init := make([]byte, 0, 32)
	dim := make([]byte, 4)
	binary.BigEndian.PutUint16(dim[0:2], uint16(w))
	binary.BigEndian.PutUint16(dim[2:4], uint16(h))
	init = append(init, dim...)
	// 像素格式：bpp=32 depth=24 big-endian=0 true-color=1 (R<<16 G<<8 B)
	init = append(init, []byte{32, 24, 0, 1, 0, 255, 0, 255, 0, 255, 16, 8, 0, 0, 0, 0}...)
	name := "OneHost VNC"
	nl := make([]byte, 4)
	binary.BigEndian.PutUint32(nl, uint32(len(name)))
	init = append(init, nl...)
	init = append(init, []byte(name)...)
	if _, err := c.Write(init); err != nil {
		return
	}
	log.Printf("VNC 客户端已连接: %s", c.RemoteAddr())

	fb := render(w, h)

	for {
		mt, err := br.ReadByte()
		if err != nil {
			return
		}
		switch mt {
		case 0: // SetPixelFormat
			if _, err := io.ReadFull(br, make([]byte, 19)); err != nil {
				return
			}
		case 2: // SetEncodings
			hdr := make([]byte, 3)
			if _, err := io.ReadFull(br, hdr); err != nil {
				return
			}
			n := int(binary.BigEndian.Uint16(hdr[1:3]))
			if _, err := io.ReadFull(br, make([]byte, n*4)); err != nil {
				return
			}
		case 3: // FramebufferUpdateRequest
			if _, err := io.ReadFull(br, make([]byte, 9)); err != nil {
				return
			}
			if err := sendFramebuffer(c, w, h, fb); err != nil {
				return
			}
		case 4: // KeyEvent
			if _, err := io.ReadFull(br, make([]byte, 7)); err != nil {
				return
			}
		case 5: // PointerEvent
			if _, err := io.ReadFull(br, make([]byte, 5)); err != nil {
				return
			}
		case 6: // ClientCutText
			hdr := make([]byte, 7)
			if _, err := io.ReadFull(br, hdr); err != nil {
				return
			}
			n := int(binary.BigEndian.Uint32(hdr[3:7]))
			if n > 0 {
				if _, err := io.ReadFull(br, make([]byte, n)); err != nil {
					return
				}
			}
		default:
			return
		}
	}
}

func sendFramebuffer(c net.Conn, w, h int, fb []byte) error {
	msg := make([]byte, 0, 16+len(fb))
	msg = append(msg, 0, 0) // type=0 FramebufferUpdate, padding
	nr := make([]byte, 2)
	binary.BigEndian.PutUint16(nr, 1)
	msg = append(msg, nr...)
	rect := make([]byte, 12)
	binary.BigEndian.PutUint16(rect[0:2], 0)
	binary.BigEndian.PutUint16(rect[2:4], 0)
	binary.BigEndian.PutUint16(rect[4:6], uint16(w))
	binary.BigEndian.PutUint16(rect[6:8], uint16(h))
	binary.BigEndian.PutUint32(rect[8:12], 0) // raw 编码
	msg = append(msg, rect...)
	msg = append(msg, fb...)
	_, err := c.Write(msg)
	return err
}

// render 生成一张品牌色渐变 + 顶部深色标题栏的帧缓冲（BGRA，小端 32 位）。
func render(w, h int) []byte {
	fb := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			var r, g, b int
			if y < 64 { // 顶部标题栏
				r, g, b = 30, 33, 48
			} else {
				// 对角渐变：靛蓝 -> 紫
				r = 60 + x*120/w
				g = 60 + y*90/h
				b = 180 + (x+y)*60/(w+h)
			}
			fb[i] = byte(b)
			fb[i+1] = byte(g)
			fb[i+2] = byte(r)
			fb[i+3] = 0
		}
	}
	return fb
}
