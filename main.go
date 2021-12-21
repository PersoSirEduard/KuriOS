package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var config map[string]interface{}
var root Folder
var currentDir *Folder

func onMessage(session *discordgo.Session, message *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if message.Author.ID == session.State.User.ID {
		return
	}

	// Ignore all messages outside the configured channel
	focus_channel := config["focus_channel"].(string)
	if message.ChannelID != focus_channel {
		return
	}

	// Ignore command if it begins with '!'
	if message.Content[0] == '!' {
		return
	}

	// Parse input command line
	commands, err := parseCommandLine(message.Content)

	if err != nil {
		fmt.Println("Error: could not parse the input command,", err)
		echo(session, message, "Error: Invalid input format", COLOR_RED)
		return
	}

	switch commands[0] {

	// Print out all available commands
	case "help":
		content, err := os.ReadFile("help.txt")

		if err != nil {
			fmt.Println("Error: Could not read the help file,", err)
			echo(session, message, "Error: Could not read the help file.", COLOR_RED)
		}

		echo(session, message, string(content), COLOR_WHITE)
		break

	// Print out a message to the output
	case "echo":
		echo(session, message, strings.Join(commands[1:], " "), COLOR_WHITE)
		break

	// Get the current directory
	case "pwd":
		echo(session, message, getCurrentDirectoryPath(), COLOR_WHITE)
		break

	case "ls":

		break

	case "cd":
		// nextDir, err := getDirectory(commands[1], true, true)

		// if err != nil {
		// 	echo(session, message, "Error: Could not find the directory.", COLOR_RED)
		// 	return
		// }

		// currentDir = nextDir.path + nextDir.name + "/"
		// echo(session, message, "Changed directory to "+getCurrentDirectoryPath(), COLOR_GREEN)
		break

	case "test":
		echo(session, message, root.path+root.name, COLOR_BLUE)

	// Unkown command
	default:
		echo(session, message, "Error: Unknown command. Use \"help\" for more information.", COLOR_RED)
		break
	}

}

func main() {

	// Get configurations
	err := getConfig(&config)

	if err != nil {
		fmt.Println("Error: Could not load configurations,", err)
		return
	}

	fmt.Println("Loaded configurations.")

	// Load the root directory
	err = loadTree()
	if err != nil {
		fmt.Println("Error: Could not load the root directory,", err)
		return
	}

	fmt.Println("Loaded directory tree.")

	// Create a new Discord session using the provided bot token.
	client, err := discordgo.New("Bot " + config["token"].(string))

	if err != nil {
		fmt.Println("Error: Could not communicate with Discord API,", err)
		return
	}

	// Listen for messages on Discord
	client.AddHandler(onMessage)

	// Give proper permissions to the bot
	client.Identify.Intents = discordgo.IntentsAll

	// Start the bot
	err = client.Open()
	if err != nil {
		fmt.Println("Error: Could not open a connection,", err)
		return
	}

	fmt.Println("KuriOS is now running.  Press CTRL-C to exit.")

	// Wait for the program to finish
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// End the bot
	client.Close()

}
