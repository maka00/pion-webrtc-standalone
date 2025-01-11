package signalling

import (
	"encoding/json"
	"fmt"
	"log"
	"pion-webrtc/internal/dto"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	id             uuid.UUID
	conn           *websocket.Conn
	senderPump     chan dto.Message
	receiverPump   chan []byte
	disposeHandler func(clnt *Client)
	msgHandler     map[string]func(message dto.Message)
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{ //nolint:exhaustruct
		conn:       conn,
		id:         uuid.New(),
		msgHandler: make(map[string]func(dto.Message)),
	}
}

func (c *Client) ID() string {
	return c.id.String()
}

func (c *Client) SetOnReceive(msgType string, onReceive func(dto.Message)) {
	c.msgHandler[msgType] = onReceive
}

func (c *Client) Start() {
	c.senderPump = make(chan dto.Message)
	c.receiverPump = make(chan []byte)

	go func() {
		for message := range c.senderPump {
			if err := c.Send(message); err != nil {
				return
			}
		}
	}()

	go func() {
		for {
			message, err := c.Receive()
			if err != nil {
				return
			}

			var msg dto.Message

			if err := json.Unmarshal(message, &msg); err != nil {
				fmt.Println("error in unmarshall")
				fmt.Printf("Message: %s\n", string(message))
			}

			if callback, ok := c.msgHandler[msg.Type]; ok {
				envelopeMessage := dto.Message{
					Type: msg.Type,
					Data: string(message),
				}
				callback(envelopeMessage)
			}
		}
	}()

	c.conn.SetCloseHandler(func(_ int, _ string) error {
		c.dispose()

		return nil
	})
}

func (c *Client) SetCloseHandler(f func(clnt *Client)) {
	c.disposeHandler = f
}

func (c *Client) dispose() {
	fmt.Println("closed")
	c.disposeHandler(c)
}

func (c *Client) Send(message dto.Message) error {
	messageBytes, err := json.Marshal(message)
	log.Printf("sending: %s\n", string(messageBytes))

	if err != nil {
		log.Println("Error marshalling message")
	}

	if err = c.conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
		return fmt.Errorf("writing error %w", err)
	}

	return nil
}

func (c *Client) Receive() ([]byte, error) {
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("receive error: %w", err)
	}

	log.Printf("receiving: %s\n", message)

	return message, nil
}
