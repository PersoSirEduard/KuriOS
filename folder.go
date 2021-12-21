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
func getCurrentDirectoryPath() string {
	return (*currentDir).path + (*currentDir).name
}

func loadTree() error {

	// Load directory tree data from directory.json
	directoryJson, err := os.Open("directory.json")

	if err != nil {
		return err
	}

	defer directoryJson.Close()
	byteValue, _ := ioutil.ReadAll(directoryJson)
	dirStruc := gjson.Parse(string(byteValue))

	// Load the root directory
	folders, files, err := _loadElement(&dirStruc, "/")

	if err != nil {
		return nil
	}

	root = Folder{"", "/", folders, files}
	currentDir = &root

	return nil
}

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

// Retrive a directory from the tree
func getDirectory(path string, formatting bool, relative bool) (*Folder, error) {

	// Support relative path formatting
	if formatting {
		err := formatPath(&path)

		if err != nil {
			return nil, err
		}
	}

	pathRoute := strings.Split(path, "/")

	currDir := &root

	if relative {
		dir, _ := getDirectory((*currentDir).path, false, false)
		currDir = dir
	}

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
