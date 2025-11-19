package ipc

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"sync"
	"time"
)

type Client struct {
	Path   string
	conn   net.Conn
	reader *bufio.Reader
	mu     sync.Mutex
}

func NewClient(path string) *Client { return &Client{Path: path} }

func (c *Client) Connect() error {
	var err error
	c.conn, err = net.Dial("unix", c.Path)
	if err != nil {
		return err
	}
	c.reader = bufio.NewReader(c.conn)
	return nil
}

func (c *Client) SendTyped(method string, payload any) (*Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		if err := c.Connect(); err != nil {
			return nil, err
		}
	}

	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	id := fmt.Sprintf("%d-%d", time.Now().UnixNano(), n.Int64())
	msg := Message{ID: id, Type: "req", Method: method, Payload: payload}

	if err := json.NewEncoder(c.conn).Encode(&msg); err != nil {
		return nil, err
	}

	var resp Message
	if err := json.NewDecoder(c.reader).Decode(&resp); err != nil {
		return nil, err
	}
	if resp.Err != "" {
		return nil, errors.New(resp.Err)
	}
	return &resp, nil
}
