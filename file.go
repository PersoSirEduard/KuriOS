package main

import (
	"errors"
	"strings"
)

// Structure of a file holding data
type File struct {
	name             string
	path             string
	data             string
	cache            string
	availableBetween []string
	locked           bool
	key              string
}

func getFile(path string, formatting bool, channel string) (*File, error) {

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
	for i, p := range pathRoute {

		// Ignore empty path
		if strings.TrimSpace(p) == "" {
			continue
		}

		if i == len(pathRoute)-1 {
			if _, okFile := (*currDir).files[p]; okFile {
				f := (*currDir).files[p]
				if !IsTimeAvailable((*f).availableBetween[0], (*f).availableBetween[1]) {
					return nil, errors.New("Error: The file is unavailable at the moment.")
				}
				if (*f).locked {
					return nil, errors.New("Error: The file is locked.")
				}
				return f, nil
			} else {
				return nil, errors.New("Error: Could not find the file " + path + ".")
			}
		}

		// Travel to the next directory
		if _, ok := (*currDir).folders[p]; ok {
			f := (*currDir).folders[p]
			if !IsTimeAvailable((*f).availableBetween[0], (*f).availableBetween[1]) {
				return nil, errors.New("Error: The directory \"" + (*f).path + (*f).name + "\" is unavailable at the moment.")
			}
			if (*f).locked {
				return nil, errors.New("Error: The directory \"" + (*f).path + (*f).name + "\" is locked.")
			}
			currDir = f

		} else {
			return nil, errors.New("Error: Could not find the file " + path + ".")
		}

	}

	return nil, errors.New("Error: Could not find the file " + path + ".")
}
