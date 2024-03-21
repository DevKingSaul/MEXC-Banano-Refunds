package util

import (
	"github.com/devkingsaul/mexc-banano-refunds/json_marshal"
	"github.com/devkingsaul/mexc-banano-refunds/uint128"
	"golang.org/x/crypto/blake2b"
	"encoding/binary"
	"slices"
)

type StateBlock struct {
	Account        Address            `json:"account"`
	Previous       json_marshal.Hex32 `json:"previous"`
	Representative Address            `json:"representative"`
	Balance        uint128.Uint128    `json:"balance"`
	Link           json_marshal.Hex32 `json:"link"`
	Signature      json_marshal.Hex64 `json:"signature"`
	Work           json_marshal.Hex8  `json:"work"`
}

func (block StateBlock) Hash() [32]byte {
	var input [176]byte
	input[31] = 6;
	copy(input[32:64], block.Account[:])
	copy(input[64:96], block.Previous[:])
	copy(input[96:128], block.Representative[:])
	block.Balance.PutBytesBE(input[128:144])
	copy(input[144:176], block.Link[:])

	return blake2b.Sum256(input[:]);
}

func (block StateBlock) WorkDifficulty() (uint64, error) {
	hash, err := blake2b.New(8, nil)

	if err != nil {
		return 0, err
	}

	var work [8]byte

	copy(work[:], block.Work[:])
	slices.Reverse(work[:])

	hash.Write(work[:])

	if IsZero(block.Previous[:]) {
		hash.Write(block.Account[:])
	} else {
		hash.Write(block.Previous[:])
	}

	threshold := hash.Sum(nil);

	return binary.LittleEndian.Uint64(threshold), nil
}