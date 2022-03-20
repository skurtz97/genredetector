package util

import (
	"encoding/json"
	"io"
)

func FromJSON[T any](r io.Reader, dest *T) error {
	err := json.NewDecoder(r).Decode(dest)
	if err != nil {
		return err
	}
	return nil
}
func ToJSON[T any](w io.Writer, src *T) error {
	err := json.NewEncoder(w).Encode(src)
	if err != nil {
		return err
	}
	return nil
}
