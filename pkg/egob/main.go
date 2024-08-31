package egob

import (
	// "encoding/json"
	"bytes"
	"encoding/gob"
)

// func Marshal(v interface{}) ([]byte, error) {
// 	return json.Marshal(v)
// }
//
// func Unmarshal(data []byte, v interface{}) error {
// 	return json.Unmarshal(data, v)
// }

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
