package block_proccessor

import (
	"encoding/json"
	"fmt"
	"github.com/devkingsaul/mexc-banano-refunds/rpc"
	"github.com/devkingsaul/mexc-banano-refunds/uint128"
	"github.com/devkingsaul/mexc-banano-refunds/util"
)

const (
	SEND_BLOCK    = 0
	RECEIVE_BLOCK = 1
)

type Block struct {
	Amount uint128.Uint128 `json:"amount"`
	Link   [32]byte        // Recepient for Send Block, Block Hash for Receive Block
}

type QueueEntry struct {
	Type  uint8 `json:"type"`
	Block Block `json:"block"`
}

type BlockController struct {
	Frontier util.StateBlock `json:"frontier"`
	Queue    []QueueEntry    `json:"queue"`
}

func Run(channel <-chan QueueEntry, frontier util.StateBlock, api rpc.APIController) {
	/*queue := make([]QueueEntry, 0, 5)

	controller := BlockController{
		Frontier: frontier,
		Queue: queue,
	}*/

	for {
		msg := <-channel

		b, err := json.MarshalIndent(msg, "", "   ")

		if err != nil {
			panic(err)
		}

		fmt.Println(string(b))
	}
}
