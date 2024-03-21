package websocket_controller

import (
	"github.com/devkingsaul/mexc-banano-refunds/block_proccessor"
	"github.com/devkingsaul/mexc-banano-refunds/json_marshal"
	"github.com/devkingsaul/mexc-banano-refunds/uint128"
	"github.com/devkingsaul/mexc-banano-refunds/util"
	"github.com/gorilla/websocket"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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

type WebSocketSubscribtion struct {
	Action string `json:"action"`
	Topic  string `json:"topic"`
	Ack    bool   `json:"ack"`
}

func ProcessMessage(rawMessage []byte) error {
	var message WebSocketMessage

	err := json.Unmarshal(rawMessage, &message)

	if err != nil {
		fmt.Println(err)
		return err
	}

	if message.Ack != "" {
		fmt.Printf("Subscribtion Acknowledgement: %s\n", message.Ack)
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

		b, err := json.MarshalIndent(topicMsg, "", "  ")

		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(b))
		}

		if topicMsg.Block.SubType == "send" {
			fmt.Println(topicMsg.Amount)
		}
	}

	return nil
}

func Start(proccessor_channel chan<- block_proccessor.QueueEntry, wsUrl string) {
	SubscribtionMessage := WebSocketSubscribtion{
		Action: "subscribe",
		Topic:  "confirmation",
		Ack:    true,
	}

	SubscribtionMessageJSON, err := json.Marshal(SubscribtionMessage)

	if err != nil {
		panic(err)
	}

ConnectionLoop:
	for {
		c, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
		if err != nil {
			fmt.Println("dial error:", err, "reconnecting...")
			time.Sleep(5 * time.Second)
			continue ConnectionLoop
		}

		fmt.Println("Connected")

		c.WriteMessage(websocket.TextMessage, SubscribtionMessageJSON)

	MessageLoop:
		for {
			mt, rawMessage, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				break MessageLoop
			}

			switch mt {
			case websocket.CloseMessage:
				break MessageLoop

			case websocket.TextMessage:
				{
					err = ProcessMessage(rawMessage)

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
