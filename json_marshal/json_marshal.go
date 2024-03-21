package json_marshal

import (
	"encoding/hex"
	"encoding/json"
	"errors"
)

type Hex8 [8]byte
type Hex32 [32]byte
type Hex64 [64]byte

func (v Hex8) MarshalJSON() ([]byte, error) {
	str := hex.EncodeToString(v[:])

	return json.Marshal(str)
}

func (v Hex32) MarshalJSON() ([]byte, error) {
	str := hex.EncodeToString(v[:])

	return json.Marshal(str)
}

func (v Hex64) MarshalJSON() ([]byte, error) {
	str := hex.EncodeToString(v[:])

	return json.Marshal(str)
}

func (v *Hex8) UnmarshalJSON(b []byte) error {
	var str string

	err := json.Unmarshal(b, &str)

	if err != nil {
		return err
	}

	value, err := hex.DecodeString(str)

	if err != nil {
		return err
	}

	if len(value) != 8 {
		return errors.New("invalid hex length")
	}

	copy(v[:], value)

	return nil
}

func (v *Hex32) UnmarshalJSON(b []byte) error {
	var str string

	err := json.Unmarshal(b, &str)

	if err != nil {
		return err
	}

	value, err := hex.DecodeString(str)

	if err != nil {
		return err
	}

	if len(value) != 32 {
		return errors.New("invalid hex length")
	}

	copy(v[:], value)

	return nil
}

func (v *Hex64) UnmarshalJSON(b []byte) error {
	var str string

	err := json.Unmarshal(b, &str)

	if err != nil {
		return err
	}

	value, err := hex.DecodeString(str)

	if err != nil {
		return err
	}

	if len(value) != 64 {
		return errors.New("invalid hex length")
	}

	copy(v[:], value)

	return nil
}
