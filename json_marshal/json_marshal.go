package json_marshal

import (
	"encoding/hex"
	"encoding/json"
)

type Hex []byte

func (v Hex) MarshalJSON() ([]byte, error) {
	str := hex.EncodeToString(v)

	return json.Marshal(str)
}

func (v *Hex) UnmarshalJSON(b []byte) error {
	var str string

	err := json.Unmarshal(b, &str)

	if err != nil {
		return err
	}

	value, err := hex.DecodeString(str)

	if err != nil {
		return err
	}

	*v = value

	return nil
}
