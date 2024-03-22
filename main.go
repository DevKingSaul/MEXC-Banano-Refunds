package main

import (
	"encoding/json"
	"fmt"
	"github.com/devkingsaul/mexc-banano-refunds/block_proccessor"
	"github.com/devkingsaul/mexc-banano-refunds/ed25519"
	"github.com/devkingsaul/mexc-banano-refunds/json_marshal"
	"github.com/devkingsaul/mexc-banano-refunds/rpc"
	"github.com/devkingsaul/mexc-banano-refunds/uint128"
	"github.com/devkingsaul/mexc-banano-refunds/util"
	"github.com/devkingsaul/mexc-banano-refunds/websocket_controller"
	"os"
)

type Config struct {
	PrivateKey        json_marshal.Hex32 `json:"private_key"`
	RefundAmount      uint128.Uint128    `json:"refund_amount"`
	WebSocket         string             `json:"ws_url"`
	RPC               string             `json:"rpc_url"`
	WithdrawalAccount util.Address       `json:"withdrawal_account"`
}

var apiController rpc.APIController
var wsController websocket_controller.WebSocketController

func main() {
	json_raw, err := os.ReadFile("./config.json")

	if err != nil {
		switch {

		case os.IsNotExist(err):
			fmt.Println("Configuration File was not found.")

		default:
			fmt.Println("Error reading Configuration File: (" + err.Error() + ")")

		}

		os.Exit(1)
	}

	var config Config

	err = json.Unmarshal(json_raw, &config)

	if err != nil {
		fmt.Println("Error parsing Configuration File: (" + err.Error() + ")")
		os.Exit(1)
	}

	privateKey := ed25519.NewKeyFromSeed(config.PrivateKey[:])
	var publicKey [32]byte

	copy(publicKey[:], privateKey[32:])

	messages := make(chan block_proccessor.QueueEntry, 5)

	apiController = rpc.APIController{Url: config.RPC}
	wsController = websocket_controller.WebSocketController{
		Proccessor:        messages,
		Url:               config.WebSocket,
		Sender:            publicKey,
		WithdrawalAccount: config.WithdrawalAccount,
		RefundAmount:      config.RefundAmount,
	}

	frontier, err := apiController.FetchAccountFrontier(publicKey)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go block_proccessor.Run(messages, frontier, apiController, privateKey)

	wsController.Start()
}
