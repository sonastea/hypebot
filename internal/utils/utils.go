package utils

import (
	"encoding/json"
	"log"
)

func CheckErrFatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func CheckErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func PrettyStruct(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}
