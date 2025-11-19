package ipc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

type Server struct {
	Path string
}

func NewServer(path string) *Server {
	return &Server{Path: path}
}

func (s *Server) Start() error {
	ln, err := net.Listen("unix", s.Path)
	if err != nil {
		return err
	}
	defer ln.Close()
	fmt.Println("IPC server listening:", s.Path)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	br := bufio.NewReader(conn)

	for {
		var msg Message
		if err := json.NewDecoder(br).Decode(&msg); err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			fmt.Println("decode error:", err)
			return
		}

		h := GetHandler(msg.Method)
		if h != nil {
			go func(m Message) {
				respPayload, err := h(m.Payload)

				// 如果是请求类型才发送响应
				if m.Type != "notify" {
					resp := Message{
						ID:      m.ID,
						Type:    "resp",
						Payload: respPayload,
					}
					if err != nil {
						resp.Err = err.Error()
					}

					// 编码并写入连接
					b, encErr := Encode(&resp)
					if encErr != nil {
						fmt.Println("encode error:", encErr)
						return
					}
					_, writeErr := conn.Write(b)
					if writeErr != nil {
						fmt.Println("write error:", writeErr)
						return
					}
				}
			}(msg)
		}
	}
}
