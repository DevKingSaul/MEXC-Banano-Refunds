package util

import (
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"slices"
	"golang.org/x/crypto/blake2b"
)

var BananoB32 = base32.NewEncoding("13456789abcdefghijkmnopqrstuwxyz")

type Address [32]byte

func (v Address) MarshalJSON() ([]byte, error) {
	str, err := EncodeAddress(v, "ban_")

	if err != nil {
		return nil, err
	}

	return json.Marshal(str)
}

func (addr *Address) UnmarshalJSON(b []byte) error {
	var str string

	err := json.Unmarshal(b, &str)

	if err != nil {
		return err
	}

	value, err := DecodeAddress(str, "ban_")

	if err != nil {
		return err
	}

	copy(addr[:], value[:])

	return nil
}

func DecodeAddress(address string, prefix string) (key Address, err error) {
	if !strings.HasPrefix(address, prefix) {
		err = errors.New("invalid address prefix")
		return
	}

	if (len(address) - len(prefix)) != 60 {
		err = errors.New("invalid address size")
		return
	}

	decoded, err := BananoB32.DecodeString("1111" + address[len(prefix):])

	if err != nil {
		return
	}

	if len(decoded) != 40 {
		err = fmt.Errorf("internal error: unexpected output length (got %d)", len(decoded))
		return
	}

	hash, err := blake2b.New(5, nil)

	if err != nil {
		return
	}

	hash.Write(decoded[3:35])

	expectedChecksum := hash.Sum(nil);

	slices.Reverse(expectedChecksum)

	if !slices.Equal(decoded[35:40], expectedChecksum) {
		err = errors.New("invalid checksum")
		return
	}

	copy(key[:], decoded[3:35])

	return
}

func EncodeAddress(key [32]byte, prefix string) (address string, err error) {
	if len(key) != 32 {
		err = errors.New("invalid key size")
		return
	}

	var input [40]byte

	hash, err := blake2b.New(5, nil)

	if err != nil {
		return
	}

	hash.Write(key[:])

	checksum := hash.Sum(nil);

	slices.Reverse(checksum)

	copy(input[3:35], key[:])
	copy(input[35:40], checksum)

	address = prefix + BananoB32.EncodeToString(input[:])[4:]

	return
}