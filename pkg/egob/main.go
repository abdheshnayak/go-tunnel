package egob

import (
	"bytes"
	"encoding/gob"
)

func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(data []byte, v interface{}) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}
