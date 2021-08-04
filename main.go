package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/canastto/api"
)

const (
	// RESPONSE is the output file name for the data
	RESPONSE = "./output.json"
)

var (
	service = api.New()
)

func main() {
	data, err := service.GetData()
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(RESPONSE, b, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
