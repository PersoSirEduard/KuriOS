package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	// System variables

	// Current version of the program
	VERSION = "0.0.1"

	// Contents of the configuration file
	// Contains discord token, focus channel, and other configurations
	config map[string]interface{}

	// Root directory folder
	root Folder
	// Current directory
	currentDir *Folder

	// System variables
	systemVariables map[string](*SystemVariable) = map[string](*SystemVariable){
		"version": &(SystemVariable{"version", true, VERSION}), // Current version of the program
	}

	// Application configuration variables
	useDiscord = flag.Bool("discord", false, "Use Discord mode")                            // Enter or not discord mode
	configPath = flag.String("config", "config.json", "Path to the configuration file")     // Path to the configuration file
	envPath    = flag.String("load", "directory.json", "Path to the environment directory") // Path to the environment directory
)

// Event handler for Discord messages
func onDiscordMessage(session *discordgo.Session, message *discordgo.MessageCreate) {

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

	// List the contents of the current directory
	case "ls":

		tree := drawDirectory(currentDir, 4)
		echo(session, message.ChannelID, tree, COLOR_WHITE)

		break

	// Change the current directory
	case "cd":

		// Check if the directory was specified
		if len(commands) < 2 {
			echo(session, message.ChannelID, "Error: No directory was specified. Expecting \"cd <directory>\".", COLOR_RED)
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
		// Check if the variable name was specified
		if len(commands) < 3 {
			echo(session, message.ChannelID, "Error: No variable name was specified. Expecting \"set <variable> <value>\".", COLOR_RED)
			return
		}

		// Check if the variable exists
		if !systemVariableExists(commands[1]) {

			// Create the variable if not found
			echo(session, message.ChannelID, "Could not already find the variable. Creating a new variable.", COLOR_YELLOW)
			err := createSystemVariable(commands[1], commands[2])
			if err != nil {
				echo(session, message.ChannelID, "Error: Could not create the variable.", COLOR_RED)
				return
			}
			echo(session, message.ChannelID, "Created variable \""+commands[1]+"\" with value \""+commands[2]+"\".\n", COLOR_GREEN)

		} else {
			// Set the value of the variable
			err := setSystemVariable(commands[1], commands[2])

			if err != nil {
				echo(session, message.ChannelID, "Error: Could not set the variable value. "+err.Error(), COLOR_RED)
				return
			}

			echo(session, message.ChannelID, "Updated variable \""+commands[1]+"\" with value \""+commands[2]+"\".\n", COLOR_GREEN)
		}
		break

	// Get a global variable
	case "get":

		// Check if the variable name was specified
		if len(commands) < 2 {
			echo(session, message.ChannelID, "Error: No variable was specified. Expecting \"get <variable>\".", COLOR_RED)
			return
		}

		// Return the value of the variable if it exists
		variable, err := getSystemVariable(commands[1])

		// Variable not found
		if err != nil {
			echo(session, message.ChannelID, err.Error(), COLOR_RED)
			return
		}

		echo(session, message.ChannelID, variable.Value, COLOR_WHITE)

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

// Event handler for command line interface
func onCliMessage(message string) {

	// Parse input command line
	commands, err := parseCommandLine(message)

	if err != nil {
		fmt.Println("Error: Invalid input format")
		return
	}

	// Execute the command
	switch commands[0] {

	// Print out all available commands
	case "help":
		content, err := os.ReadFile("help.txt")

		if err != nil {
			fmt.Println("Error: Could not read the help file.")
		}

		fmt.Println(string(content))
		break

	// Print out a message to the output
	case "echo":
		fmt.Println(strings.Join(commands[1:], " "))
		break

	// Get the current directory
	case "pwd":
		fmt.Println(getCurrentDirectoryPath())
		break

	// List the contents of the current directory
	case "ls":

		tree := drawDirectory(currentDir, 4)
		fmt.Println(tree)

		break

	// Change the current directory
	case "cd":

		// Check if the directory was specified
		if len(commands) < 2 {
			fmt.Println("Error: No directory was specified. Expecting \"cd <directory>\".")
			return
		}

		nextDir, err := getDirectory(commands[1], true)

		if err != nil {
			fmt.Println("Error: Could not find the directory.")
			return
		}

		currentDir = nextDir
		fmt.Println("Changed directory to " + getCurrentDirectoryPath())
		break

	// Print the contents of a file
	case "cat":
		break

	// Download the contents of a file (for discord only)
	case "grab":
		break

	// Set a global variable
	case "set":
		// Check if the variable name was specified
		if len(commands) < 3 {
			fmt.Println("Error: No variable name or value was specified. Expecting \"set <variable> <value>\".")
			return
		}

		// Check if the variable exists
		if !systemVariableExists(commands[1]) {

			// Create the variable if not found
			fmt.Println("Could not already find the variable. Creating a new variable.")
			err := createSystemVariable(commands[1], commands[2])
			if err != nil {
				fmt.Println("Error: Could not create the variable.")
				return
			}
			fmt.Printf("Created variable \"%s\" with value \"%s\".\n", commands[1], commands[2])

		} else {
			// Set the value of the variable
			err := setSystemVariable(commands[1], commands[2])

			if err != nil {
				fmt.Println("Error: Could not set the variable value.", err)
				return
			}

			fmt.Printf("Updated variable \"%s\" with value \"%s\".\n", commands[1], commands[2])
		}
		break

	// Get a global variable
	case "get":
		// Check if the variable name was specified
		if len(commands) < 2 {
			fmt.Println("Error: No variable was specified. Expecting \"get <variable>\".")
			return
		}

		// Return the value of the variable if it exists
		variable, err := getSystemVariable(commands[1])

		// Variable not found
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Println(variable.Value)
		break

	// Delete a global variable
	case "delete":
		break

	// Unkown command
	default:
		fmt.Println("Error: Unknown command. Use \"help\" for more information.")
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
		fmt.Println("Error: Could not load configurations.", err)
		return
	}

	// Load the root directory and environment
	fmt.Println("Loaded configurations.")
	err = loadEnvironment(*envPath)
	if err != nil {
		fmt.Println("Error: Could not load the environment.", err)
		return
	}
	fmt.Println("Loaded the environment.")

	// Create a new Discord session using the provided bot token in the config file
	if *useDiscord {

		fmt.Println("Using Discord mode.")

		// Create a new Discord session using the provided bot token.
		client, err := discordgo.New("Bot " + config["token"].(string))
		if err != nil {
			fmt.Println("Error: Could not communicate with Discord API.", err)
			return
		}

		// Listen for messages on Discord
		client.AddHandler(onDiscordMessage)

		// Give proper permissions to the bot
		client.Identify.Intents = discordgo.IntentsAll

		// Start the bot
		err = client.Open()
		if err != nil {
			fmt.Println("Error: Could not open a connection.", err)
			return
		}

		fmt.Printf("KuriOS v%s is now running. Press CTRL-C to exit.", VERSION)

		// Wait for the program to finish
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
		<-sc

		// End the bot
		client.Close()

	} else {

		// Create a new command line interface
		fmt.Println("Using CLI mode.")

		// Initialize the command line interface reader
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("KuriOS v%s is now running. Type \"help\" for more information.\n", VERSION)

		// Exit the program when the user presses CTRL-C
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		go func() {
			<-sc
			fmt.Println("\nExiting...")
			os.Exit(1)
		}()

		// Listen for messages on the command line
		for {
			fmt.Print("> ")

			// Read the command line input and execute it
			command, _ := reader.ReadString('\n')
			onCliMessage(command)
			fmt.Print("\n") // Skip a line
		}
	}

}
