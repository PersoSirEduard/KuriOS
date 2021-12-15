package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Supported color constants
const (
	COLOR_WHITE  = "\n"
	COLOR_RED    = "diff\n-"
	COLOR_GREEN  = "diff\n+"
	COLOR_BLUE   = "yaml\n"
	COLOR_YELLOW = "fix\n"
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

// Parse a string into a list of individual commands
func parseCommandLine(input string) ([]string, error) {

	// Read file and properly parse the input
	reader := csv.NewReader(strings.NewReader(input))
	reader.Comma = ' '
	fields, err := reader.Read()

	// Check for any input errors
	if err != nil {
		return nil, err
	}

	return fields, nil
}

// Return a message back to the user
func echo(session *discordgo.Session, discordMsg *discordgo.MessageCreate, message string, color string) {
	_, err := session.ChannelMessageSend(discordMsg.ChannelID, "```"+color+message+"```")

	if err != nil {
		fmt.Println("Error: Could not send a response message,", err)
		echo(session, discordMsg, "Error: Exceeded allowed output text length.", COLOR_RED)
	}
}
