package utils

import "encoding/json"

func PrettyStruct(data any) (string, error) {
	s, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(s), nil
}
