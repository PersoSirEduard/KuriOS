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

// Supported color constants for discord output
const (
	COLOR_WHITE  = "\n"
	COLOR_RED    = "diff\n-"
	COLOR_GREEN  = "diff\n+"
	COLOR_BLUE   = "yaml\n"
	COLOR_YELLOW = "fix\n"
)

// Load configurations from a specified file and insert them into a config map
// @param config: *map[string]interface{} - The configuration map that will be loaded
// @param path: string - The path to the configuration file
// @return error - Any error that may have occurred
func getConfig(config *map[string]interface{}, path string) error {

	// Load the configuration file
	jsonConfig, err := os.Open(path)
	if err != nil {
		return err
	}

	// Close the file
	defer jsonConfig.Close()

	// Read the file's contents
	byteValue, _ := ioutil.ReadAll(jsonConfig)

	// Convert the byte array to a string and JSON parse it
	json.Unmarshal(byteValue, config)

	return nil
}

// Parse an incoming command and return the command and its arguments
// @param message: string - The command to parse
// @return []string - The command parsed
// @return error - Any error that may have occurred
func parseCommandLine(input string) ([]string, error) {

	// Read the command and properly parse the input
	reader := csv.NewReader(strings.NewReader(input))
	reader.Comma = ' '
	fields, err := reader.Read()

	return fields, err
}

// **DISCORD FEATURE ONLY**
// Print a message to the discord channel
// @param session: *discordgo.Session - The discord session to use
// @param channelId: string - The channel to print the message to
// @param message: string - The message to print
// @param color: string - The color of the message
func echo(session *discordgo.Session, channelId string, message string, color string) {

	// Send the message to the discord channel
	_, err := session.ChannelMessageSend(channelId, "```"+color+message+"```")
	if err != nil {
		fmt.Println("Error: Could not send a response message,", err)
		echo(session, channelId, "Error: Exceeded allowed output text length.", COLOR_RED)
	}
}

// Support for file and folder directory formatting
// @param path: *string - The path to the directory
// @return error - Any error that may have occurred
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
