package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

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
	root = Folder{"", "/", folders, files}
	currentDir = &root

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

		if elType == "folder" {

			// Load subfolder
			children := value.Get("children")
			subFolder, subFiles, err := _loadElement(&children, path+name+"/")

			if err != nil {
				return nil, nil, err
			}

			folders[name] = &Folder{name, path, subFolder, subFiles}

		} else if elType == "file" {

			// Load file
			subFile := File{name, path, nil, ""}

			// Check to see if file has cache
			if value.Get("cache").Exists() {
				subFile.cache = value.Get("cache").String()
			}

			// Check to see if file has data
			if value.Get("data").Exists() {
				subFile.data = []byte(value.Get("data").String())
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

	// Handle the directory structure to save
	outputEnvStruct := "\"struct\": {"     // Start the directory structure
	folderCount := len(currentDir.folders) // Count the number of folders remaining
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
	output += "\"children\": {"                // Start the children

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
