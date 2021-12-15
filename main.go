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
	focus_channel := fmt.Sprintf("%d", int(config["focus_channel"].(float64)))
	if strings.Compare(message.ChannelID, focus_channel) != 1 {
		return
	}

	session.ChannelMessageSend(message.ChannelID, "Hello, "+message.Author.Username+"!")
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
