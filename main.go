package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	// System variables

	// Current version of the program
	VERSION = "0.0.1"

	// Max number of channels
	MAX_CHANNELS = 5

	// Contents of the configuration file
	// Contains discord token, focus channel, and other configurations
	config map[string]interface{}

	// Root directory folder
	root Folder

	// Open channels
	openChannels []string

	// Current directory
	currentDir map[string](*Folder)

	// System variables
	systemVariables map[string](*SystemVariable) = map[string](*SystemVariable){
		"version": &(SystemVariable{"version", true, VERSION}), // Current version of the program
		"time":    &(SystemVariable{"time", false, "0"}),       // Current time
	}

	// User permissions
	systemPermissions   map[string]([]string) = map[string]([]string){}
	priorityPermissions []string              = []string{}

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

	// Ignore all messages outside the configured channels
	isInChannel := false
	for _, channel := range config["focus_channels"].([]interface{}) {
		if channel.(string) == message.ChannelID {
			isInChannel = true
			// Check to see if the key is part of the currentDir
			// If not, create it
			if _, ok := currentDir[message.ChannelID]; !ok {
				currentDir[message.ChannelID] = &root
			}
			break
		}
	}

	for i := 0; i < len(openChannels); i++ {
		if openChannels[i] == message.ChannelID {
			isInChannel = true
			break
		}
	}

	if !isInChannel {
		return
	}

	// Ignore command if it begins with '!'
	if len(message.Content) == 0 {
		return
	}
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

		if !HasPermission(session, message.Member, message.GuildID, "help") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		content, err := os.ReadFile("help.txt")

		if err != nil {
			fmt.Println("Error: Could not read the help file,", err)
			echo(session, message.ChannelID, "Error: Could not read the help file.", COLOR_RED)
		}

		echo(session, message.ChannelID, string(content), COLOR_WHITE)
		break

	// Print out a message to the output
	case "echo":

		if !HasPermission(session, message.Member, message.GuildID, "echo") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		echo(session, message.ChannelID, strings.Join(commands[1:], " "), COLOR_WHITE)
		break

	// Get the current directory
	case "pwd":

		if !HasPermission(session, message.Member, message.GuildID, "pwd") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		echo(session, message.ChannelID, getCurrentDirectoryPath(message.ChannelID), COLOR_WHITE)
		break

	// List the contents of the current directory
	case "ls":

		if !HasPermission(session, message.Member, message.GuildID, "ls") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		tree := drawDirectory(currentDir[message.ChannelID], 4)
		echo(session, message.ChannelID, tree, COLOR_WHITE)

		break

	// Change the current directory
	case "cd":

		if !HasPermission(session, message.Member, message.GuildID, "cd") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		// Check if the directory was specified
		if len(commands) < 2 {
			echo(session, message.ChannelID, "Error: No directory was specified. Expecting \"cd <directory>\".", COLOR_RED)
			return
		}

		nextDir, err := getDirectory(commands[1], true, message.ChannelID)

		if err != nil {
			echo(session, message.ChannelID, err.Error(), COLOR_RED)
			return
		}

		currentDir[message.ChannelID] = nextDir
		echo(session, message.ChannelID, "Changed directory to "+getCurrentDirectoryPath(message.ChannelID), COLOR_GREEN)
		break

	// Print the contents of a file
	case "cat":

		if !HasPermission(session, message.Member, message.GuildID, "cat") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		// Get the directory
		if len(commands) < 2 {
			echo(session, message.ChannelID, "Error: No file was specified. Expecting \"grab <file>\".", COLOR_RED)
			return
		}

		file, err := getFile(commands[1], true, message.ChannelID)
		if err != nil {
			echo(session, message.ChannelID, err.Error(), COLOR_RED)
			return
		}

		// Get the file contents
		if (*file).cache != "" {

			// Get the file contents from the cache
			data, err := os.ReadFile((*file).cache)
			if err != nil {
				echo(session, message.ChannelID, "Error: Could not open the file \""+(*file).cache+"\".", COLOR_RED)
				return
			}

			// Convert the file contents to a string
			content := string(data)
			// Send the file contents to the output
			echo(session, message.ChannelID, content, COLOR_WHITE)

		} else {
			// Read the file contents from the data
			content := (*file).data
			// Send the file contents to the output
			echo(session, message.ChannelID, content, COLOR_WHITE)
		}

		break

	// Download the contents of a file (for discord only)
	case "grab":

		if !HasPermission(session, message.Member, message.GuildID, "grab") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		// Get the directory
		if len(commands) < 2 {
			echo(session, message.ChannelID, "Error: No file was specified. Expecting \"grab <file>\".", COLOR_RED)
			return
		}

		file, err := getFile(commands[1], true, message.ChannelID)
		if err != nil {
			echo(session, message.ChannelID, "Error: Could not find the file \""+commands[1]+"\".", COLOR_RED)
			return
		}

		// Get the file contents
		if (*file).cache != "" {
			// Load from cache folder the file
			// Check if the file exists

			data, err := os.ReadFile((*file).cache)
			if err != nil {
				echo(session, message.ChannelID, "Error: Could not open the file \""+(*file).cache+"\".", COLOR_RED)
				return
			}

			// Create an io reader from the file's bytes
			reader := bytes.NewReader(data)

			// Send the file to the user
			session.ChannelFileSend(message.ChannelID, (*file).name, reader)
		} else {
			// Create an io reader from a string
			reader := strings.NewReader((*file).data)

			// Send the file to the user
			session.ChannelFileSend(message.ChannelID, (*file).name, reader)
		}

		break

	// Set a global variable
	case "set":

		if !HasPermission(session, message.Member, message.GuildID, "set") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

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

		if !HasPermission(session, message.Member, message.GuildID, "get") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

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

		if !HasPermission(session, message.Member, message.GuildID, "delete") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		err = deleteSystemVariable(commands[1])
		if err != nil {
			echo(session, message.ChannelID, err.Error(), COLOR_RED)
			return
		}

		echo(session, message.ChannelID, "Deleted variable \""+commands[1]+"\".", COLOR_GREEN)

	case "su":

		if !HasPermission(session, message.Member, message.GuildID, "su") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		if len(commands) < 3 {
			echo(session, message.ChannelID, "Error: Incomplete command. Expecting \"su <subscribe | unsubscribe> <role>\".", COLOR_RED)
			return
		}

		switch commands[1] {
		case "subscribe":
			err := subscribeToRole(session, message.Member, message.GuildID, message.Author.ID, commands[2])
			if err != nil {
				echo(session, message.ChannelID, err.Error(), COLOR_RED)
				return
			}
			echo(session, message.ChannelID, "Subscribed to role \""+commands[2]+"\".", COLOR_GREEN)
			break
		case "unsubscribe":
			err := unsubscribeFromRole(session, message.Member, message.GuildID, message.Author.ID, commands[2])
			if err != nil {
				echo(session, message.ChannelID, err.Error(), COLOR_RED)
				return
			}
			echo(session, message.ChannelID, "Unsubscribed from role \""+commands[2]+"\".", COLOR_GREEN)
			break
		default:
			echo(session, message.ChannelID, "Error: Invalid command. Expecting \"su <subscribe | unsubscribe> <role>\".", COLOR_RED)
		}
		break

	case "lock":

		if !HasPermission(session, message.Member, message.GuildID, "lock") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		if len(commands) < 3 {
			echo(session, message.ChannelID, "Error: Incomplete command. Expecting \"lock <path> <key>\".", COLOR_RED)
			return
		}

		file, err := getFile(commands[1], true, message.ChannelID)
		if err != nil {
			folder, err := getDirectory(commands[1], true, message.ChannelID)
			if err != nil {
				echo(session, message.ChannelID, "Error: Cannot access or find the specified path.", COLOR_RED)
				return
			}
			err = LockFolder(folder, commands[2])
			if err != nil {
				echo(session, message.ChannelID, err.Error(), COLOR_RED)
				return
			}
		} else {
			err = LockFile(file, commands[2])
			if err != nil {
				echo(session, message.ChannelID, err.Error(), COLOR_RED)
				return
			}
		}

		// Redirect to the root
		currentDir[message.ChannelID] = &root

		echo(session, message.ChannelID, "Locked \""+commands[1]+"\" and redirected to root.", COLOR_GREEN)
		break

	case "unlock":

		if !HasPermission(session, message.Member, message.GuildID, "unlock") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		if len(commands) < 3 {
			echo(session, message.ChannelID, "Error: Incomplete command. Expecting \"unlock <path> <key>\".", COLOR_RED)
			return
		}

		file, err := getFile(commands[1], true, message.ChannelID)
		if err != nil && file == nil {
			folder, err := getDirectory(commands[1], true, message.ChannelID)
			if err != nil && folder == nil {
				echo(session, message.ChannelID, "Error: Cannot access or find the specified path.", COLOR_RED)
				return
			}
			err = UnlockFolder(folder, commands[2])
			if err != nil {
				echo(session, message.ChannelID, err.Error(), COLOR_RED)
				return
			}
		} else {
			err = UnlockFile(file, commands[2])
			if err != nil {
				echo(session, message.ChannelID, err.Error(), COLOR_RED)
				return
			}
		}

		echo(session, message.ChannelID, "Unlocked \""+commands[1]+"\".", COLOR_GREEN)
		break

	case "save":

		if !HasPermission(session, message.Member, message.GuildID, "save") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
		}

		if len(commands) < 2 {
			fmt.Println("Error: No file name was specified. Expecting \"save <file>\".")
			return
		}

		err := saveEnvironment(commands[1])
		if err != nil {
			echo(session, message.ChannelID, err.Error(), COLOR_RED)
		}

		echo(session, message.ChannelID, "Saved environment to file \""+commands[1]+"\".", COLOR_GREEN)
		break

	case "load":

		if !HasPermission(session, message.Member, message.GuildID, "load") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
		}

		if len(commands) < 2 {
			echo(session, message.ChannelID, "Error: No file name was specified. Expecting \"load <file>\".", COLOR_RED)
		}

		err := loadEnvironment(commands[1])
		if err != nil {
			echo(session, message.ChannelID, err.Error(), COLOR_RED)
		}

		echo(session, message.ChannelID, "Loaded environment from file \""+commands[1]+"\".", COLOR_GREEN)
		break

	case "new":

		if !HasPermission(session, message.Member, message.GuildID, "new") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
		}

		// Check amount of channels active
		if len(openChannels) >= MAX_CHANNELS {
			echo(session, message.ChannelID, "Error: Too many channels open. Please close some before creating a new one.", COLOR_RED)
			return
		}

		// Create a new channel on Discord
		channelName := NewChannelName(5)
		channel, err := session.GuildChannelCreate(message.GuildID, channelName, discordgo.ChannelTypeGuildText)
		if err != nil {
			echo(session, message.ChannelID, "Error: Failed to create channel \""+channelName+"\".", COLOR_RED)
			return
		}

		// Get parent channel
		parentChannel := ""
		channels, _ := session.GuildChannels(message.GuildID)
		for i := 0; i < len(channels); i++ {
			if channels[i].ID == message.ChannelID {
				parentChannel = channels[i].ParentID
				break
			}
		}

		// Move the channel
		_, err = session.ChannelEditComplex(channel.ID, &discordgo.ChannelEdit{ParentID: parentChannel})
		if err != nil {
			println(err.Error())
			echo(session, message.ChannelID, "Error: Failed to edit channel \""+channelName+"\".", COLOR_RED)
			return
		}

		currentDir[channel.ID] = &root
		openChannels = append(openChannels, channel.ID)
		echo(session, message.ChannelID, "Created new channel \""+channelName+"\".", COLOR_GREEN)
		echo(session, channel.ID, "New channel created. Type \"close\" to end the channel.", COLOR_WHITE)

		break

	case "close":

		if !HasPermission(session, message.Member, message.GuildID, "close") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		// Check if the channel can be closed
		// Check if it is a focus channel
		canClose := true
		for _, channel := range config["focus_channels"].([]interface{}) {
			if channel.(string) == message.ChannelID {
				canClose = false
			}
		}
		if !canClose {
			echo(session, message.ChannelID, "Error: Cannot close this channel.", COLOR_RED)
			return
		}

		// Remove channel from currentDir
		delete(currentDir, message.ChannelID)

		// Remove channel from openChannels
		for i := 0; i < len(openChannels); i++ {
			if openChannels[i] == message.ChannelID {
				openChannels = append(openChannels[:i], openChannels[i+1:]...)
				break
			}
		}

		// Delete channel
		_, err := session.ChannelDelete(message.ChannelID)
		if err != nil {
			echo(session, message.ChannelID, "Error: Failed to delete channel.", COLOR_RED)
			return
		}
		break

	case "run":

		if !HasPermission(session, message.Member, message.GuildID, "run") {
			echo(session, message.ChannelID, "Error: You do not have permission to use this command", COLOR_RED)
			return
		}

		// Get the first attachment
		if len(message.Attachments) < 1 {
			echo(session, message.ChannelID, "Error: No Kode attachment was found.", COLOR_RED)
			return
		}

		attachment := message.Attachments[0]

		// Download the attachment
		resp, err := http.Get(attachment.URL)
		if err != nil {
			echo(session, message.ChannelID, "Error: Failed to download attachment.", COLOR_RED)
			return
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			echo(session, message.ChannelID, "Error: Failed to read attachment.", COLOR_RED)
			return
		}

		body = []byte(string(body) + "\nexit\n")

		// Run the code
		cmd := exec.Command("kode", "-runStdIn")
		stdin, err := cmd.StdinPipe()

		if err != nil {
			echo(session, message.ChannelID, "Error: Failed to create stdin pipe.", COLOR_RED)
		}

		stdout, err := cmd.StdoutPipe()

		if err != nil {
			echo(session, message.ChannelID, "Error: Failed to create stdout pipe.", COLOR_RED)
		}

		scanner := bufio.NewScanner(stdout)
		done := make(chan bool)

		go func() {
			defer stdin.Close()

			if _, err := stdin.Write(body); err != nil {
				echo(session, message.ChannelID, "Error: Failed to write to stdin.", COLOR_RED)
			}

			for scanner.Scan() {
				if scanner.Text() == "exit" {
					done <- true
				}
				echo(session, message.ChannelID, scanner.Text(), COLOR_WHITE)
			}
			done <- true
		}()

		cmd.Start()
		<-done
		err = cmd.Wait()

		if err != nil {
			// println("Kode: " + err.Error())
			if err.Error() != "exit status 1" && err.Error() != "exit status 0xc000013a" {
				echo(session, message.ChannelID, "Error: Program crashed.", COLOR_RED)
			}
		}

		echo(session, message.ChannelID, "Done.", COLOR_GREEN)

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
		fmt.Println(getCurrentDirectoryPath("default"))
		break

	// List the contents of the current directory
	case "ls":

		tree := drawDirectory(currentDir["default"], 4)
		fmt.Println(tree)

		break

	// Change the current directory
	case "cd":

		// Check if the directory was specified
		if len(commands) < 2 {
			fmt.Println("Error: No directory was specified. Expecting \"cd <directory>\".")
			return
		}

		nextDir, err := getDirectory(commands[1], true, "default")

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		currentDir["default"] = nextDir
		fmt.Println("Changed directory to " + getCurrentDirectoryPath("default"))
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

		err = deleteSystemVariable(commands[1])
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("Deleted variable \"%s\".\n", commands[1])
		break

	case "save":

		if len(commands) < 2 {
			fmt.Println("Error: No file name was specified. Expecting \"save <file>\".")
			return
		}

		err := saveEnvironment(commands[1])
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println("Saved environment to file \"" + commands[1] + "\".")
		break

	case "load":

		if len(commands) < 2 {
			fmt.Println("Error: No file name was specified. Expecting \"load <file>\".")
		}

		err := loadEnvironment(commands[1])
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println("Loaded environment from file \"" + commands[1] + "\".")
		break

	case "lock":

		if len(commands) < 3 {
			fmt.Println("Error: Incomplete command. Expecting \"lock <path> <key>\".")
			return
		}

		file, err := getFile(commands[1], true, "default")
		if err != nil {
			folder, err := getDirectory(commands[1], true, "default")
			if err != nil {
				fmt.Println("Error: Cannot access or find the specified path.")
				return
			}
			err = LockFolder(folder, commands[2])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		} else {
			err = LockFile(file, commands[2])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		fmt.Println("Locked \"" + commands[1] + "\".")
		break

	case "unlock":

		if len(commands) < 3 {
			fmt.Println("Error: Incomplete command. Expecting \"unlock <path> <key>\".")
			return
		}

		file, err := getFile(commands[1], true, "default")
		if err != nil {
			folder, err := getDirectory(commands[1], true, "default")
			if err != nil {
				fmt.Println("Error: Cannot access or find the specified path.")
				return
			}
			err = UnlockFolder(folder, commands[2])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		} else {
			err = UnlockFile(file, commands[2])
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

		fmt.Println("Unlocked \"" + commands[1] + "\".")
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

		fmt.Printf("KuriOS v%s is now running. Press CTRL-C to exit.\n", VERSION)

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
