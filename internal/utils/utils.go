package utils

import (
	"encoding/json"
	"log"
)

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func PrettyStruct(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}
