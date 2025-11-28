package obs

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	OpIdentify = 1
	OpRequest  = 6

	RequestTypeStartRecord = "StartRecord"
	RequestTypeStopRecord  = "StopRecord"

	DefaultWriteDeadline = 15 * time.Second
)

type OBSHello struct {
	Op int `json:"op"`
	D  struct {
		ObsWebSocketVersion string `json:"obsWebSocketVersion"`
		RpcVersion          int
		Authentication      struct {
			Challenge string
			Salt      string
		}
	}
}

type OBSIdentify struct {
	Op int `json:"op"`
	D  struct {
		RpcVersion     int    `json:"rpcVersion"`
		Authentication string `json:"authentication"`
	} `json:"d"`
}

type OBSRequest struct {
	Op int             `json:"op"`
	D  *OBSRequestData `json:"d"`
}

type OBSRequestData struct {
	RequestType string `json:"requestType"`
	RequestId   string `json:"requestId"`
}

type Client struct {
	addr     string
	password string
	conn     *websocket.Conn
	mu       sync.Mutex
}

func NewClient(addr, password string) *Client {
	return &Client{
		addr:     addr,
		password: password,
	}
}

func (c *Client) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.addr, nil)
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}
	c.conn = conn

	// 1. Hello受信
	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read hello error: %w", err)
	}
	var hello OBSHello
	if err := json.Unmarshal(msg, &hello); err != nil {
		return fmt.Errorf("unmarshal hello error: %w", err)
	}

	// 2. Identify送信
	auth := c.makeAuth(c.password, hello.D.Authentication.Salt, hello.D.Authentication.Challenge)

	identify := OBSIdentify{
		Op: OpIdentify,
		D: struct {
			RpcVersion     int    `json:"rpcVersion"`
			Authentication string `json:"authentication"`
		}{
			RpcVersion:     hello.D.RpcVersion,
			Authentication: auth,
		},
	}

	c.mu.Lock()
	if err := c.conn.WriteJSON(identify); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("write identify error: %w", err)
	}
	c.mu.Unlock()

	// 3. Identifyレスポンス待機
	_, msg, err = c.conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read identified error: %w", err)
	}
	slog.Info("Connected to OBS", "identify_response", string(msg))

	return nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	return c.conn.Close()
}

func (c *Client) StartRecording() error {
	req := OBSRequest{
		Op: OpRequest,
		D: &OBSRequestData{
			RequestType: RequestTypeStartRecord,
			RequestId:   "start1",
		},
	}
	return c.writeRequest(req)
}

func (c *Client) StopRecording() error {
	req := OBSRequest{
		Op: OpRequest,
		D: &OBSRequestData{
			RequestType: RequestTypeStopRecord,
			RequestId:   "stop1",
		},
	}
	return c.writeRequest(req)
}

func (c *Client) writeRequest(req OBSRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection is not established")
	}

	_ = c.conn.SetWriteDeadline(time.Now().Add(DefaultWriteDeadline))
	if err := c.conn.WriteJSON(req); err != nil {
		return fmt.Errorf("write json error: %w", err)
	}
	return nil
}

func (c *Client) makeAuth(password, salt, challenge string) string {
	h1 := sha256.Sum256([]byte(password + salt))
	secret := base64.StdEncoding.EncodeToString(h1[:])
	h2 := sha256.Sum256([]byte(secret + challenge))
	return base64.StdEncoding.EncodeToString(h2[:])
}
