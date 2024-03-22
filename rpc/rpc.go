package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devkingsaul/mexc-banano-refunds/json_marshal"
	"github.com/devkingsaul/mexc-banano-refunds/util"
	"net/http"
	"slices"
)

type BlockHistoryResponse struct {
	util.StateBlock
	Hash json_marshal.Hex32 `json:"hash"`
}

type HistoryResponse struct {
	History []BlockHistoryResponse `json:"history"`
	Error   string                 `json:"error"`
}

type HistoryRequest struct {
	Action  string `json:"action"`
	Account string `json:"account"`
	Count   string `json:"count"`
	Raw     bool   `json:"raw"`
}

type APIController struct {
	Url        string `json:"rpc_url"`
	httpClient http.Client
}

type ProcessBlock struct {
	Type string `json:"type"`
	util.StateBlock
	Work string `json:"work,omitempty"`
}

type ProcessRequest struct {
	Action  string       `json:"action"`
	IsJson  bool         `json:"json_block"`
	SubType string       `json:"subtype"`
	DoWork  bool         `json:"do_work"`
	Block   ProcessBlock `json:"block"`
}

type ProcessResponse struct {
	Hash    json_marshal.Hex32 `json:"hash"`
	Error   string             `json:"error"`
}

func (controller APIController) ProcessBlock(block util.StateBlock, subtype string) (hash [32]byte, err error) {
	req := ProcessRequest{
		Action:  "process",
		IsJson:  true,
		SubType: subtype,
		DoWork:  true,
		Block: ProcessBlock{
			Type:       "state",
			StateBlock: block,
		},
	}

	rawReq, err := json.Marshal(req)

	if err != nil {
		return
	}

	resp, err := controller.httpClient.Post(controller.Url, "application/json", bytes.NewBuffer(rawReq))

	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("received bad status code (got %d)", resp.StatusCode)
		return
	}

	var respBody ProcessResponse

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	resp.Body.Close()

	if err != nil {
		return
	}

	if respBody.Error != "" {
		err = errors.New(respBody.Error)
		return
	}

	expectedHash := block.Hash()

	if !slices.Equal(expectedHash[:], respBody.Hash[:]) {
		err = errors.New("invalid block hash")
		return
	}

	copy(hash[:], respBody.Hash[:])

	return
}

func (controller APIController) FetchAccountFrontier(account [32]byte) (block util.StateBlock, err error) {
	encodedAddress, err := util.EncodeAddress(account, "ban_")

	if err != nil {
		return
	}

	req := HistoryRequest{
		Action:  "account_history",
		Account: encodedAddress,
		Count:   "1",
		Raw:     true,
	}

	rawReq, err := json.Marshal(req)

	if err != nil {
		return
	}

	resp, err := controller.httpClient.Post(controller.Url, "application/json", bytes.NewBuffer(rawReq))

	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("received bad status code (got %d)", resp.StatusCode)
		return
	}

	var respBody HistoryResponse

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	resp.Body.Close()

	if err != nil {
		return
	}

	if respBody.Error != "" {
		err = errors.New(respBody.Error)
		return
	}

	if len(respBody.History) < 1 {
		err = errors.New("no frontier")
		return
	}

	block = respBody.History[0].StateBlock

	copy(block.Account[:], account[:])

	expectedHash := block.Hash()

	if !slices.Equal(expectedHash[:], respBody.History[0].Hash[:]) {
		err = errors.New("invalid block hash")
		return
	}

	return
}
