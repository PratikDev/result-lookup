package utils

import "encoding/json"

func ToJSON[T any](model T) (string, error) {
	bytes, err := json.Marshal(model)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}