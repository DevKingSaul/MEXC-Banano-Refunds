package rpc

import (
	"github.com/devkingsaul/mexc-banano-refunds/json_marshal"
	"github.com/devkingsaul/mexc-banano-refunds/util"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
