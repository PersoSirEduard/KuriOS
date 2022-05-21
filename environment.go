package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/tidwall/gjson"
)

// Load a specified environment into memory
// @param path: string - The environment path to load
// @return error - The error if any
func loadEnvironment(path string) error {

	// Load directory tree data from the specified path
	directoryJson, err := os.Open(path)
	if err != nil {
		return err
	}

	// Close the file
	defer directoryJson.Close()

	// Read the file's contents
	byteValue, _ := ioutil.ReadAll(directoryJson)
	// Convert the byte array to a string and JSON parse it
	dirStruc := gjson.Parse(string(byteValue))

	// Clear variables
	systemVariables = map[string](*SystemVariable){
		"version": &(SystemVariable{"version", true, VERSION}), // Current version of the program
		"time":    &(SystemVariable{"time", false, "0"}),       // Current time
	}
	// Load the variables if they exist
	variables := dirStruc.Get("vars")
	if variables.Exists() {
		// Load each variable to the environment
		for name, value := range variables.Map() {
			err := createSystemVariable(name, value.String())

			// Check to see if the variable was created properly
			if err != nil {
				fmt.Printf("Exception encountered while loading variable \"%s\". Skipping it.\n", name)
			}
		}
	}

	// Clear permissions
	systemPermissions = make(map[string]([]string))
	priorityPermissions = []string{}
	// Load permissions if they exist
	permissions := dirStruc.Get("perms")
	if permissions.Exists() {
		// Load each permission to the environment
		for name, value := range permissions.Map() {
			priorityPermissions = append([]string{name}, priorityPermissions...)
			// Create the permission array
			array := make([]string, 0)
			for _, element := range value.Array() {
				array = append(array, element.String())
			}

			systemPermissions[name] = array
		}
	}

	// Access the directory structure element from the JSON
	dirStruc = dirStruc.Get("struct")
	if !dirStruc.Exists() {
		return errors.New("Could not find the directory structure. The environment file might be corrupted or invalid.")
	}

	// Load the root directory
	folders, files, err := _loadElement(&dirStruc, "/")
	if err != nil {
		return nil
	}

	// Create and set the root directory
	root = Folder{"", "/", folders, files, []string{"*", "*"}, false, ""}

	// Reset the current directory
	currentDir = map[string]*Folder{}
	currentDir["default"] = &root

	return nil
}

// Load inner environment data into memory
// @param element: *gjson.Result - Element to explore and load into memory
// @param path: string - The path to the element
// @return map[string]*Folder - The folders in the element
// @return map[string]*File - The files in the element
func _loadElement(element *gjson.Result, path string) (map[string]*Folder, map[string]*File, error) {

	// Initialize the folder and file maps
	folders := make(map[string]*Folder)
	files := make(map[string]*File)

	// Load the files and folders
	for name, value := range (*element).Map() {

		// Get sub element and its type
		elType := value.Get("type").String()

		availableBetween := []string{"*", "*"}
		if value.Get("availableBetween").Exists() {
			availableBetween = []string{}
			for _, element := range value.Get("availableBetween").Array() {

				iTime := element.String()
				// Check to see if the time is valid format
				_, err := time.Parse("2006-01-02 15:04:05", iTime)
				if err != nil {
					fmt.Printf("Exception encountered while loading time \"%s\". Invalid time format. Expected YYYY-MM-DD HH:MM:SS. Setting default value \"*\".\n", iTime)
				}

				// Add the time to the available between array
				availableBetween = append(availableBetween, element.String())
			}

			// Check if only two elements are available
			if len(availableBetween) != 2 {
				fmt.Printf("Exception encountered while loading folder \"%s\". Invalid availableBetween argument. Setting default value [\"*\", \"*\"].\n", path+name)
				availableBetween = []string{"*", "*"}
			}
			// ? Check if the first element is greater than the second

		}

		locked := false
		lockKey := ""
		if value.Get("locked").Exists() {
			locked = value.Get("locked").Bool()
			if locked {
				if value.Get("key").Exists() {
					lockKey = value.Get("key").String()
				} else {
					fmt.Printf("Exception encountered while loading folder \"%s\". Invalid lock key. Setting default value \"admin\".\n", path+name)
				}
			}
		}

		if elType == "folder" {

			// Load subfolder
			children := value.Get("children")
			subFolder, subFiles, err := _loadElement(&children, path+name+"/")

			if err != nil {
				return nil, nil, err
			}

			folders[name] = &Folder{name, path, subFolder, subFiles, availableBetween, locked, lockKey}

		} else if elType == "file" {

			// Load file
			subFile := File{name, path, "", "", availableBetween, locked, lockKey}

			// Check to see if file has cache
			if value.Get("cache").Exists() {
				subFile.cache = value.Get("cache").String()
			} else if value.Get("data").Exists() {
				// Check to see if file has data
				subFile.data = value.Get("data").String()
			}

			files[name] = &subFile

		} else {
			return nil, nil, errors.New("Unknown element type for " + path + name)
		}

	}

	return folders, files, nil
}

// Save the current environment to a specified path
// @param path: string - The path to save the environment to
// @return error - The error if any
func saveEnvironment(path string) error {

	// Handle variables to save
	outputEnvVar := "\"vars\": {"    // Start the variables
	varCount := len(systemVariables) // Count the number of variables remaining
	for _, variable := range systemVariables {

		// Ignore immutable variables
		if !(*variable).Immutable {

			// Add a comma if not the last element
			if varCount > 1 {
				outputEnvVar += "\"" + (*variable).Name + "\": \"" + (*variable).Value + "\","
			} else {
				outputEnvVar += "\"" + (*variable).Name + "\": \"" + (*variable).Value + "\""
			}
		}

		// Decrement the variable count
		varCount--
	}
	outputEnvVar += "}," // Close the variables

	// Handle permissions to save
	outputEnvPerm := "\"perms\": {"     // Start the permissions
	permCount := len(systemPermissions) // Count the number of permissions remaining
	for name, permissions := range systemPermissions {
		outputEnvPerm += "\"" + name + "\": ["

		for i, perm := range permissions {
			if i != len(permissions)-1 {
				outputEnvPerm += "\"" + perm + "\","
			} else {
				outputEnvPerm += "\"" + perm + "\""
			}
		}

		if permCount > 1 {
			outputEnvPerm += "],"
		} else {
			outputEnvPerm += "]"
		}

		// Decrement the permission count
		permCount--
	}

	// Handle the directory structure to save
	outputEnvStruct := "\"struct\": {" // Start the directory structure
	folderCount := len(root.folders)   // Count the number of folders remaining
	for _, folder := range root.folders {

		// Add a comma if not the last element
		if folderCount > 1 {
			outputEnvStruct += _saveElement(folder) + ","
		} else {
			outputEnvStruct += _saveElement(folder) + ""
		}

		folderCount-- // Decrement the folder count
	}
	outputEnvStruct += "}" // Close the directory structure

	// Concate the variables and directory structure
	outputEnv := "{" + outputEnvVar + outputEnvStruct + "}"

	// Write the environment to the specified path
	file, err := os.Create(path)

	if err != nil {
		return errors.New("Could not save the environment file.")
	}

	_, err = file.WriteString(outputEnv)

	if err != nil {
		return errors.New("Could not save the environment file.")
	}

	return nil
}

// Save an element to a specified path
// @param element: *Folder - The element to save
// @return string - The string representation of the element
func _saveElement(current *Folder) string {

	output := "\"" + (*current).name + "\": {" // Start the folder
	output += "\"type\": \"folder\","          // Set the type to folder
	if (*current).availableBetween[0] != "*" && (*current).availableBetween[1] != "*" {
		output += "\"availableBetween\": [\"" + (*current).availableBetween[0] + "\", \"" + (*current).availableBetween[1] + "\"],"
	}
	if (*current).locked {
		output += "\"locked\": true,"
		output += "\"key\": \"" + (*current).key + "\","
	}
	output += "\"children\": {" // Start the children

	folderCount := len((*current).folders) // Count the number of folders remaining
	fileCount := len((*current).files)     // Count the number of files remaining

	for _, folder := range (*current).folders {

		if folderCount > 1 || fileCount > 0 {
			// Add comma if not the last element or if there are files
			output += _saveElement(folder) + ","
		} else {
			output += _saveElement(folder) + ""
		}

		folderCount-- // Decrement the folder remaining count
	}

	for _, file := range (*current).files {
		output += "\"" + file.name + "\": {" // Start the file
		output += "\"type\": \"file\","      // Set the type to file
		if file.cache != "" {
			output += "\"cache\": \"" + file.cache + "\"," // Set the cache to the file
		}
		output += "\"data\": \"" + string(file.data) + "\"" // Set the data to the file
		output += "}"                                       // End the file

		fileCount-- // Decrement the file remaining count
	}

	output += "}}" // close children and close the folder

	return output

}
