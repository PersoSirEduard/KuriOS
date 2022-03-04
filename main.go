package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Initialization arguments
var (
	useDiscord = flag.Bool("discord", false, "Use Discord")                                 // Enter or not discord mode
	configPath = flag.String("config", "config.json", "Path to the configuration file")     // Path to the configuration file
	envPath    = flag.String("load", "directory.json", "Path to the environment directory") // Path to the environment directory
)

// Contents of the configuration file
// Contains discord token, focus channel, and other configurations
var config map[string]interface{}

// Root directory folder
var root Folder

// Current directory
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
		echo(session, message.ChannelID, "Error: Invalid input format", COLOR_RED)
		return
	}

	// Execute the command
	switch commands[0] {

	// Print out all available commands
	case "help":
		content, err := os.ReadFile("help.txt")

		if err != nil {
			fmt.Println("Error: Could not read the help file,", err)
			echo(session, message.ChannelID, "Error: Could not read the help file.", COLOR_RED)
		}

		echo(session, message.ChannelID, string(content), COLOR_WHITE)
		break

	// Print out a message to the output
	case "echo":
		echo(session, message.ChannelID, strings.Join(commands[1:], " "), COLOR_WHITE)
		break

	// Get the current directory
	case "pwd":
		echo(session, message.ChannelID, getCurrentDirectoryPath(), COLOR_WHITE)
		break

	case "ls":

		break

	// Change the current directory
	case "cd":

		// Check if the directory was specified
		if len(commands) < 2 {
			echo(session, message.ChannelID, "Error: No directory was specified.", COLOR_RED)
			return
		}

		nextDir, err := getDirectory(commands[1], true)

		if err != nil {
			echo(session, message.ChannelID, "Error: Could not find the directory.", COLOR_RED)
			return
		}

		currentDir = nextDir
		echo(session, message.ChannelID, "Changed directory to "+getCurrentDirectoryPath(), COLOR_GREEN)
		break

	// Print the contents of a file
	case "cat":
		break

	// Download the contents of a file (for discord only)
	case "grab":
		break

	// Set a global variable
	case "set":
		break

	// Get a global variable
	case "get":
		break

	// Delete a global variable
	case "delete":
		break

	// Unkown command
	default:
		echo(session, message.ChannelID, "Error: Unknown command. Use \"help\" for more information.", COLOR_RED)
		break
	}

}

// Main execution
func main() {

	// Parse the command line arguments at initialization
	// If the -discord flag is set, use discord mode
	// If the -load flag is set, use the specified directory
	// If the -config flag is set, use the specified configuration file
	flag.Parse()

	// Get configurations from the configPath
	err := getConfig(&config, *configPath)
	if err != nil {
		fmt.Println("Error: Could not load configurations,", err)
		return
	}

	// Load the root directory and environment
	fmt.Println("Loaded configurations.")
	err = loadEnvironment(*envPath)
	if err != nil {
		fmt.Println("Error: Could not load environment,", err)
		return
	}
	fmt.Println("Loaded the environment.")

	// Create a new Discord session using the provided bot token in the config file
	if *useDiscord {
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

		fmt.Println("KuriOS is now running. Press CTRL-C to exit.")

		// Wait for the program to finish
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc

		// End the bot
		client.Close()

	} else {

		println("KuriOS currently only supports Discord mode. Use the -discord flag instead. The program will now exit.")

	}

}
