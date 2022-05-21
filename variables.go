package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Structure of a system variable holding data
type SystemVariable struct {
	Name      string
	Immutable bool
	Value     string
}

// If the system variable exists, return it, otherwise return nil
// @param name: string - The name of the system variable
// @return *SystemVariable - The system variable
// @return error - The error if the system variable does not exist
func getSystemVariable(name string) (*SystemVariable, error) {
	if systemVariableExists(name) {

		if name == "time" {
			timeInt, _ := strconv.ParseInt(systemVariables["time"].Value, 10, 64)
			// Format the int unix to a string date
			return &SystemVariable{"time", false, time.Unix(time.Now().Unix()-timeInt, 0).Format("2006-01-02 15:04:05")}, nil
		}

		return systemVariables[name], nil
	} else {
		return nil, errors.New("Error: The variable \"" + name + "\" does not exist.")
	}
}

// Set a system variable to a specified value, if the variable exists
// @param name: string - The name of the system variable to set
// @param value: string - The value of the system variable
// @return error - The error if the system variable does not exist or is immutable
func setSystemVariable(name string, value string) error {

	// Check to see if the variable exists
	// If it does, get the variable
	variable, err := getSystemVariable(name)
	if err != nil {
		return err
	}

	// Check to see if the variable can be changed
	if !(*variable).Immutable {

		// Special case for time
		if (*variable).Name == "time" {

			if value == "now" {
				(*variable).Value = fmt.Sprint(0)
			} else {
				// Parse date time format string to unix time

				newTime, err := time.Parse("2006-01-02 15:04:05", value)

				if err != nil {
					return errors.New("Error: The time \"" + value + "\" is not in the correct format \"YYYY-MM-DD HH:MM:SS\".")
				}

				deltaTime := time.Now().Unix() - newTime.Unix()

				systemVariables["time"].Value = fmt.Sprint(deltaTime)
			}

			return nil

		}

		// Set the variable to the specified value
		(*variable).Value = value
		return nil
	} else {
		return errors.New("Error: The variable \"" + name + "\" is immutable.")
	}
}

// Create a system variable if it does not exist
// @param name: string - The name of the system variable to create
// @param value: string - The value of the system variable
// @return error - The error if the system variable already exists
func createSystemVariable(name string, value string) error {
	if systemVariableExists(name) {
		return errors.New("Error: The variable \"" + name + "\" already exists.")
	} else {
		systemVariables[name] = &SystemVariable{name, false, value}
		return nil
	}
}

// Determine if a system variable exists already
// @param name: string - The name of the system variable
// @return bool - True if the system variable exists, false otherwise
func systemVariableExists(name string) bool {
	_, ok := systemVariables[name]
	return ok
}

// Delete a system variable if it exists
// @param name: string - The name of the system variable to delete
// @return error - The error if the system variable does not exist
func deleteSystemVariable(name string) error {
	if systemVariableExists(name) {

		// Check status of the variable
		variable, _ := getSystemVariable(name)
		if (*variable).Immutable || (*variable).Name == "time" {
			return errors.New("Error: Cannot delete the immutable or essential system variable \"" + name + "\".")
		}

		// Remove the system variable from the map
		delete(systemVariables, name)
		return nil
	} else {
		return errors.New("Error: The variable \"" + name + "\" does not exist.")
	}
}
