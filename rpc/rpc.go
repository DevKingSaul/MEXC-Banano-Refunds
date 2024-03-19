package rpc

import "net/http"

type AccountInfo struct {
	Frontier [32]byte
	OpenBlock [32]byte
	
}

type AccountInfoJSON struct {
	Frontier string `json:"frontier"`
}

func RPC_AccountInfo() {

}
