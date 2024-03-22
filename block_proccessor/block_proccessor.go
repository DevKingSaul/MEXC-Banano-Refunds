package block_proccessor

import (
	"fmt"
	"github.com/devkingsaul/mexc-banano-refunds/ed25519"
	"github.com/devkingsaul/mexc-banano-refunds/rpc"
	"github.com/devkingsaul/mexc-banano-refunds/uint128"
	"github.com/devkingsaul/mexc-banano-refunds/util"
	"strings"
	"encoding/hex"
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

func Run(channel <-chan QueueEntry, frontier util.StateBlock, api rpc.APIController, privateKey []byte) {
	/*queue := make([]QueueEntry, 0, 5)

	controller := BlockController{
		Frontier: frontier,
		Queue: queue,
	}*/

	for {
		msg := <-channel

		var newBalance uint128.Uint128
		var subtype string

		frontierHash := frontier.Hash()

		if msg.Type == SEND_BLOCK {
			newBalance = frontier.Balance.Sub(msg.Block.Amount)
			subtype = "send"

			recepientAddr, err := util.EncodeAddress(msg.Block.Link, "ban_")

			if err != nil {
				fmt.Printf("Failed to parse Address (%s)", err)
				continue
			}

			fmt.Printf("Sending\n  Frontier: %s\n  Recepient: %s\n  Amount: %s\n", strings.ToUpper(hex.EncodeToString(frontierHash[:])), recepientAddr, msg.Block.Amount.String())
		} else if msg.Type == RECEIVE_BLOCK {
			newBalance = frontier.Balance.Add(msg.Block.Amount)
			subtype = "receive"

			fmt.Printf("Receiving\n  Frontier: %s\n  Link: %s\n  Amount: %s\n", strings.ToUpper(hex.EncodeToString(frontierHash[:])), strings.ToUpper(hex.EncodeToString(msg.Block.Link[:])), msg.Block.Amount.String())
		} else {
			fmt.Printf("Received invaild Block Type (%d)\n", msg.Type)
			continue
		}

		newBlock := util.StateBlock{
			Account:        frontier.Account,
			Previous:       frontierHash,
			Representative: frontier.Representative,
			Balance:        newBalance,
			Link:           msg.Block.Link,
		}

		blockHash := newBlock.Hash()

		signature := ed25519.Sign(privateKey, blockHash[:])

		copy(newBlock.Signature[:], signature)

		_, err := api.ProcessBlock(newBlock, subtype)

		if err != nil {
			fmt.Printf("Received error when processing block (%s)\n", err)
		} else {
			frontier = newBlock
		}
	}
}
