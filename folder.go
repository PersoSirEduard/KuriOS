package main

import (
	"errors"
	"strings"
)

// Structure containing other directories and files
type Folder struct {
	name             string
	path             string
	folders          map[string]*Folder
	files            map[string]*File
	availableBetween []string
	locked           bool
	key              string
}

// Get the current working directory
// @return string - The current working directory
func getCurrentDirectoryPath(channel string) string {
	return (*(currentDir)[channel]).path + (*(currentDir)[channel]).name
}

// Retrive a directory folder
// @param path: string - The path to the directory
// @param formatting: bool - Whether or not formatting syntax is used
// @return *Folder - The directory folder
func getDirectory(path string, formatting bool, channel string) (*Folder, error) {

	// Check to see if the specified path is the root
	if path == "/" {
		return &root, nil
	}

	// Support relative path formatting
	if formatting {
		err := formatPath(&path, channel)

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
			f := (*currDir).folders[p]
			if !IsTimeAvailable((*f).availableBetween[0], (*f).availableBetween[1]) {
				return nil, errors.New("Error: The directory is unavailable at the moment.")
			}
			if (*f).locked {
				return f, errors.New("Error: The directory \"" + (*f).path + (*f).name + "\" is locked.")
			}
			currDir = f

		} else {
			return nil, errors.New("Error: Could not find the directory.")
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

		timeAvailable := IsTimeAvailable((*folder).availableBetween[0], (*folder).availableBetween[1])

		// Draw sub folder names
		if countFolder == numOfFolders-1 && numOfFiles == 0 {
			// Reached the last folder

			if timeAvailable && !(*folder).locked {
				// Not locked and time available
				dirTree = append(dirTree, prefix+"└─"+folder.name+"/")
			} else if !timeAvailable && !(*folder).locked {
				// Not locked and time not available
				dirTree = append(dirTree, prefix+"└─"+folder.name+"/ (unavailable)")
			} else {
				// Locked and time available or not available
				dirTree = append(dirTree, prefix+"└─"+folder.name+"/ (locked)")
			}

			// Draw sub folder contents
			if timeAvailable && !(*folder).locked {
				content := drawDirectoryUtil(folder, prefix+"  ", depth-1)
				dirTree = append(dirTree, content...)
			}

		} else {

			if timeAvailable && !(*folder).locked {
				// Not locked and time available
				dirTree = append(dirTree, prefix+"├─"+folder.name+"/")
			} else if !timeAvailable && !(*folder).locked {
				// Not locked and time not available
				dirTree = append(dirTree, prefix+"├─"+folder.name+"/ (unavailable)")
			} else {
				// Locked and time available or not available
				dirTree = append(dirTree, prefix+"├─"+folder.name+"/ (locked)")
			}

			// Draw sub folder contents
			if timeAvailable && !(*folder).locked {
				content := drawDirectoryUtil(folder, prefix+"| ", depth-1)
				dirTree = append(dirTree, content...)
			}

		}

		countFolder++
	}

	// Draw the files
	for _, file := range directory.files {

		timeAvailable := IsTimeAvailable((*file).availableBetween[0], (*file).availableBetween[1])

		// Draw file names
		if countFile == numOfFiles-1 {
			// Reached the last file
			if timeAvailable && !(*file).locked {
				// Not locked and time available
				dirTree = append(dirTree, prefix+"└─"+file.name)
			} else if !timeAvailable && !(*file).locked {
				// Not locked and time not available
				dirTree = append(dirTree, prefix+"└─"+file.name+" (unavailable)")
			} else {
				// Locked and time available or not available
				dirTree = append(dirTree, prefix+"└─"+file.name+" (locked)")
			}

		} else {
			if timeAvailable && !(*file).locked {
				// Not locked and time available
				dirTree = append(dirTree, prefix+"├─"+file.name)
			} else if !timeAvailable && !(*file).locked {
				// Not locked and time not available
				dirTree = append(dirTree, prefix+"├─"+file.name+" (unavailable)")
			} else {
				// Locked and time available or not available
				dirTree = append(dirTree, prefix+"├─"+file.name+" (locked)")
			}
		}

		countFile++
	}

	return dirTree
}
