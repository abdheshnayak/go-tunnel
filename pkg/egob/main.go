package egob

import (
	"encoding/json"
	"fmt"
	// "bytes"
	// "encoding/gob"
)

func Marshal(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s: %w", v, err)
	}
	return b, nil
}

func Unmarshal(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal %q: %w", string(data), err)
	}
	return nil
}

// func Marshal(v interface{}) ([]byte, error) {
// 	var buf bytes.Buffer
// 	enc := gob.NewEncoder(&buf)
// 	err := enc.Encode(v)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }
//
// func Unmarshal(data []byte, v interface{}) error {
// 	dec := gob.NewDecoder(bytes.NewReader(data))
// 	return dec.Decode(v)
// }
