package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Load configurations from config.json
func getConfig(config *map[string]interface{}) error {

	jsonConfig, err := os.Open("config.json")

	if err != nil {
		return err
	}

	defer jsonConfig.Close()
	byteValue, _ := ioutil.ReadAll(jsonConfig)
	json.Unmarshal(byteValue, config)

	return nil
}
