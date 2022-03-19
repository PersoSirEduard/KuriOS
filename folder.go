package main

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

// Structure containing other directories and files
type Folder struct {
	name    string
	path    string
	folders map[string]*Folder
	files   map[string]*File
}

// Get the current working directory
// @return string - The current working directory
func getCurrentDirectoryPath() string {
	return (*currentDir).path + (*currentDir).name
}

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

// Retrive a directory folder
// @param path: string - The path to the directory
// @param formatting: bool - Whether or not formatting syntax is used
// @return *Folder - The directory folder
func getDirectory(path string, formatting bool) (*Folder, error) {

	// Check to see if the specified path is the root
	if path == "/" {
		return &root, nil
	}

	// Support relative path formatting
	if formatting {
		err := formatPath(&path)

		if err != nil {
			return nil, err
		}
	}

	currDir := &root

	// Get travel steps
	pathRoute := strings.Split(path, "/")

	// Navigate through the path
	for _, p := range pathRoute {

		// Ignore empty path
		if strings.TrimSpace(p) == "" {
			continue
		}

		// Travel to the next directory
		if _, ok := (*currDir).folders[p]; ok {
			currDir = (*currDir).folders[p]

		} else {
			return nil, errors.New("Could not find directory " + path)
		}

	}

	return currDir, nil
}

// Draw the directory tree (only the first level)
// @param directory: *Folder - The directory to draw
// @param maxDepth: int - Max tree depth to draw
// @return string - The directory tree
func drawDirectory(directory *Folder, maxDepth int) string {

	// Initialize the tree name
	treeOutput := directory.name + "/\n|\n"

	// Generate the tree recursively
	tree := drawDirectoryUtil(directory, "", maxDepth)
	for _, line := range tree {
		treeOutput += line + "\n"
	}

	return treeOutput
}

// Utility function to draw the directory tree recursively
// @param directory: *Folder - The directory to draw
// @param prefix: string - The prefix to use before each line of the tree
// @param depth: int - The current depth of the tree. At depth 0, the discovery of the directory is done.
// @return []string - The directory tree as a list of lines
func drawDirectoryUtil(directory *Folder, prefix string, depth int) []string {

	// Get the size of the directory
	numOfFolders := len(directory.folders)
	countFolder := 0
	numOfFiles := len(directory.files)
	countFile := 0

	dirTree := make([]string, 0)

	// Check if the max discovery depth has been reached
	if depth < 0 {

		// Check if there are any items remaining
		// If there are no items remaining, then we are at the end of the tree
		// Otherwise, we show that there are more items
		if (numOfFolders + numOfFiles) > 0 {
			dirTree = append(dirTree, prefix+"└─(...)")
		}

		return dirTree
	}

	// Draw the folders
	for _, folder := range directory.folders {

		// Draw sub folder names
		if countFolder == numOfFolders-1 && numOfFiles == 0 {
			// Reached the last folder

			dirTree = append(dirTree, prefix+"└─"+folder.name+"/")

			// Draw sub folder contents
			content := drawDirectoryUtil(folder, prefix+"  ", depth-1)
			dirTree = append(dirTree, content...)

		} else {

			dirTree = append(dirTree, prefix+"├─"+folder.name+"/")

			// Draw sub folder contents
			content := drawDirectoryUtil(folder, prefix+"| ", depth-1)
			dirTree = append(dirTree, content...)

		}

		countFolder++
	}

	// Draw the files
	for _, file := range directory.files {

		// Draw file names
		if countFile == numOfFiles-1 {
			// Reached the last file
			dirTree = append(dirTree, prefix+"└─"+file.name)
		} else {
			dirTree = append(dirTree, prefix+"├─"+file.name)
		}

		countFile++
	}

	return dirTree
}
