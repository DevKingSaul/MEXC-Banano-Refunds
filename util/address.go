package util

import (
	"encoding/base32"
)

var BananoB32 = base32.NewEncoding("13456789abcdefghijkmnopqrstuwxyz")

type AddressMarshal [32]byte
