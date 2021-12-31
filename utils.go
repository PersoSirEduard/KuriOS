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

func formatPath(path *string) error {

	// Return to parent directory
	if strings.Contains(*path, "..") {

		currPath := getCurrentDirectoryPath()

		for _, dir := range strings.Split(*path, "/") {

			if dir == ".." {
				currDir, err := getDirectory(currPath, false)

				if err != nil {
					return err
				}

				currPath = (*currDir).path

			} else if dir != "" {
				currPath += "/" + dir
			}
		}

		*path = currPath
		return nil

	} else if (*path)[0] == '~' {
		// Return to root directory
		*path = "/" + (*path)[1:]
		return nil
	} else if (*path)[0] == '.' {
		// Return to current directory
		*path = getCurrentDirectoryPath() + "/" + (*path)[1:]
		return nil
	} else {
		*path = getCurrentDirectoryPath() + "/" + *path
		return nil
	}
}
