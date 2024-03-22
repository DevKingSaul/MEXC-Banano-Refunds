package websocket_controller

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devkingsaul/mexc-banano-refunds/block_proccessor"
	"github.com/devkingsaul/mexc-banano-refunds/json_marshal"
	"github.com/devkingsaul/mexc-banano-refunds/uint128"
	"github.com/devkingsaul/mexc-banano-refunds/util"
	"github.com/gorilla/websocket"
	"slices"
	"time"
)

type block struct {
	util.StateBlock
	SubType string `json:"subtype"`
}

type ConfirmationTopic struct {
	Account util.Address       `json:"account"`
	Hash    json_marshal.Hex32 `json:"hash"`
	Amount  uint128.Uint128    `json:"amount"`
	Block   block              `json:"block"`
}

type WebSocketMessage struct {
	Ack     string          `json:"ack,omitempty"`
	Topic   string          `json:"topic,omitempty"`
	Message json.RawMessage `json:"message,omitempty"`
}

type SubscriptionOptions struct {
	Accounts []util.Address `json:"accounts"`
}

type WebSocketSubscription struct {
	Action  string              `json:"action"`
	Topic   string              `json:"topic"`
	Ack     bool                `json:"ack"`
	Options SubscriptionOptions `json:"options"`
}

type WebSocketController struct {
	Proccessor        chan<- block_proccessor.QueueEntry
	Url               string
	Sender            util.Address
	WithdrawalAccount util.Address
	RefundAmount      uint128.Uint128
	ProccessedHashes  map[[32]byte]bool
}

func (controller WebSocketController) ProcessMessage(rawMessage []byte) error {
	var message WebSocketMessage

	err := json.Unmarshal(rawMessage, &message)

	if err != nil {
		fmt.Println(err)
		return err
	}

	if message.Ack != "" {
		fmt.Printf("Subscription Acknowledgement: %s\n", message.Ack)
	} else if message.Topic == "confirmation" {
		var topicMsg ConfirmationTopic

		err = json.Unmarshal(message.Message, &topicMsg)

		if err != nil {
			fmt.Println(err)
			return err
		}

		hash := topicMsg.Block.Hash()

		if !slices.Equal(hash[:], topicMsg.Hash[:]) {
			fmt.Printf("Invalid Block Hash: %s, %s\n", hex.EncodeToString(hash[:]), hex.EncodeToString(topicMsg.Hash[:]))
			return errors.New("invalid hash")
		}

		if topicMsg.Block.SubType == "send" {
			_, exists := controller.ProccessedHashes[hash]
			if exists {
				return nil
			}

			controller.ProccessedHashes[hash] = true

			if slices.Equal(topicMsg.Block.Link[:], controller.Sender[:]) {
				controller.Proccessor <- block_proccessor.QueueEntry{
					Type: block_proccessor.RECEIVE_BLOCK,
					Block: block_proccessor.Block{
						Amount: topicMsg.Amount,
						Link:   hash,
					},
				}
			} else if slices.Equal(topicMsg.Block.Account[:], controller.WithdrawalAccount[:]) {
				controller.Proccessor <- block_proccessor.QueueEntry{
					Type: block_proccessor.SEND_BLOCK,
					Block: block_proccessor.Block{
						Amount: controller.RefundAmount,
						Link:   topicMsg.Block.Link,
					},
				}
			}
		}
	}

	return nil
}

func (controller WebSocketController) Start() {
	controller.ProccessedHashes = make(map[[32]byte]bool)

	SubscriptionMessage := WebSocketSubscription{
		Action: "subscribe",
		Topic:  "confirmation",
		Ack:    true,
		Options: SubscriptionOptions{
			Accounts: []util.Address{controller.Sender, controller.WithdrawalAccount},
		},
	}

	SubscriptionMessageJSON, err := json.Marshal(SubscriptionMessage)

	if err != nil {
		panic(err)
	}

ConnectionLoop:
	for {
		c, _, err := websocket.DefaultDialer.Dial(controller.Url, nil)
		if err != nil {
			fmt.Println("dial error:", err, "reconnecting...")
			time.Sleep(5 * time.Second)
			continue ConnectionLoop
		}

		fmt.Println("Connected")

		c.WriteMessage(websocket.TextMessage, SubscriptionMessageJSON)

	MessageLoop:
		for {
			eventType, rawMessage, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				break MessageLoop
			}

			switch eventType {
			case websocket.CloseMessage:
				break MessageLoop

			case websocket.TextMessage:
				{
					err = controller.ProcessMessage(rawMessage)

					if err != nil {
						break MessageLoop
					}
				}
			}
		}

		c.Close()

		fmt.Println("disconnected reconnecting...")
		time.Sleep(5 * time.Second)
	}
}
