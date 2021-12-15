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

	// Parse input command line
	commands, err := parseCommandLine(message.Content)

	if err != nil {
		fmt.Println("Error: could not parse the input command,", err)
		echo(session, message, "Error: Invalid input format", COLOR_RED)
		return
	}

	switch commands[0] {
	case "help":
		content, err := os.ReadFile("help.txt")

		if err != nil {
			fmt.Println("Error: Could not read the help file,", err)
			echo(session, message, "Error: Could not read the help file.", COLOR_RED)
		}

		echo(session, message, string(content), COLOR_WHITE)
		break

	case "echo":
		echo(session, message, strings.Join(commands[1:], " "), COLOR_WHITE)
		break

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
