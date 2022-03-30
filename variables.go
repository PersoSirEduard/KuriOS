package main

import "errors"

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
		return systemVariables[name], nil
	} else {
		return nil, errors.New("Error: The variable does not exist.")
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
		// Set the variable to the specified value
		(*variable).Value = value
		return nil
	} else {
		return errors.New("Error: The variable is immutable.")
	}
}

// Create a system variable if it does not exist
// @param name: string - The name of the system variable to create
// @param value: string - The value of the system variable
// @return error - The error if the system variable already exists
func createSystemVariable(name string, value string) error {
	if systemVariableExists(name) {
		return errors.New("Error: The variable already exists.")
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
