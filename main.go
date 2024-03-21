package main

import (
	"github.com/devkingsaul/mexc-banano-refunds/ed25519"
	"github.com/devkingsaul/mexc-banano-refunds/json_marshal"
	"github.com/devkingsaul/mexc-banano-refunds/rpc"
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	PrivateKey   json_marshal.Hex32 `json:"private_key"`
	RefundAmount float32            `json:"refund_amount"`
	WebSocket    string             `json:"ws_url"`
	RPC          string             `json:"rpc_url"`
}

var apiController rpc.APIController

func main() {
	json_raw, err := os.ReadFile("./config.json")

	if err != nil {
		switch {

		case os.IsNotExist(err):
			fmt.Println("Configuration File was not found.")

		default:
			fmt.Println("Error reading Configuration File: (" + err.Error() + ")")

		}

		return
	}

	var config Config

	err = json.Unmarshal(json_raw, &config)

	if err != nil {
		fmt.Println("Error parsing Configuration File: (" + err.Error() + ")")

		return
	}

	privateKey := ed25519.NewKeyFromSeed(config.PrivateKey[:])
	var publicKey [32]byte

	copy(publicKey[:], privateKey[32:])

	apiController = rpc.APIController{Url: config.RPC}

	fmt.Println(config)

	_, err = apiController.FetchAccountFrontier(publicKey)

	if err != nil {
		fmt.Println(err)
	}

}
