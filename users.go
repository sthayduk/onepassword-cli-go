package onepassword

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// UserType represents the type of a user.
type UserType string

const (
	UserTypeMember UserType = "MEMBER"
)

// UserState represents the state of a user.
type UserState string

const (
	UserStateActive            UserState = "ACTIVE"
	UserStateTransferStarted   UserState = "TRANSFER_STARTED"
	UserStateSuspended         UserState = "SUSPENDED"
	UserStateTransferSuspended UserState = "TRANSFER_SUSPENDED"
)

// User represents a user in the 1Password system.
type User struct {
	cli *OpCLI `json:"-"` // Reference to the OpCLI instance for update operations

	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Type       UserType  `json:"type"`
	State      UserState `json:"state"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastAuthAt time.Time `json:"last_auth_at"`
}

// ListUsers retrieves a list of all users in the 1Password system.
// It executes the "op user list" command using the OpCLI instance.
//
// Returns:
// - A slice of User objects representing the users in the system.
// - An error if the command execution or JSON unmarshalling fails.
func (cli *OpCLI) ListUsers() ([]User, error) {

	// Execute the command to list users
	output, err := cli.ExecuteOpCommand("user", "list")
	if err != nil {
		return nil, err
	}

	var users []User
	err = json.Unmarshal([]byte(output), &users)
	if err != nil {
		return nil, err
	}
	// Set the cli reference for each user
	for i := range users {
		users[i].cli = cli
	}

	return users, nil
}

func (cli *OpCLI) getUser(userID string) (*User, error) {
	// Execute the command to get a user by ID
	output, err := cli.ExecuteOpCommand("user", "get ", userID)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal([]byte(output), &user)
	if err != nil {
		return nil, err
	}

	user.cli = cli

	return &user, nil
}

// GetUserByName retrieves a user by their name.
// It uses the "op user get" command to fetch the user details.
//
// Parameters:
// - userName: The name of the user to retrieve.
//
// Returns:
// - A pointer to the User object if found.
// - An error if the user is not found or the command fails.
func (cli *OpCLI) GetUserByName(userName string) (*User, error) {
	return cli.getUser(userName)
}

func (cli *OpCLI) GetUserByEmail(userEmail string) (*User, error) {
	// Validate the email format
	if !isValidEmail(userEmail) {
		return nil, fmt.Errorf("invalid email format: %s", userEmail)
	}

	return cli.getUser(userEmail)
}

func (cli *OpCLI) GetUserByID(userID string) (*User, error) {
	return cli.getUser(userID)
}

// ProvisionUser creates a new user in the 1Password system.
// It uses the "op user provision" command to create the user.
//
// Parameters:
// - name: The name of the user to create.
// - email: The email address of the user.
// - language: The preferred language of the user (default is "en").
//
// Returns:
// - A pointer to the newly created User object.
// - An error if the command fails or the email format is invalid.
func (cli *OpCLI) ProvisionUser(name, email, language string) (*User, error) {
	// Validate the email format
	if !isValidEmail(email) {
		return nil, fmt.Errorf("invalid email format: %s", email)
	}

	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	if language == "" {
		language = "en"
	}

	// Execute the command to provision a new user
	output, err := cli.ExecuteOpCommand("user", "provision", "--name", name, "--email", email, "--language", language)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal([]byte(output), &user)
	if err != nil {
		return nil, err
	}

	user.cli = cli

	return &user, nil
}

// isValidEmail validates if a given string is a valid email address.
func isValidEmail(email string) bool {
	// Define a regular expression for validating email addresses.
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(emailRegex)
	return regex.MatchString(email)
}

// Confirm confirms a user by their ID using the 1Password CLI.
// It executes the "user confirm" command with the user's ID and parses
// the resulting output into an updated User object.
//
// Returns:
//   - A pointer to the updated User object if the confirmation is successful.
//   - An error if the command execution or JSON unmarshalling fails.
func (user *User) Confirm() (*User, error) {
	// Execute the command to confirm a user by ID
	output, err := user.cli.ExecuteOpCommand("user", "confirm", user.ID)
	if err != nil {
		return nil, err
	}

	var updatedUser User
	err = json.Unmarshal([]byte(output), &updatedUser)
	if err != nil {
		return nil, err
	}

	updatedUser.cli = user.cli

	return &updatedUser, nil
}

// Delete removes a user from the 1Password system.
// It uses the "op user delete" command to delete the user by their ID.
//
// Returns:
// - An error if the command fails.
func (user *User) Delete() error {
	// Execute the command to delete a user by ID
	_, err := user.cli.ExecuteOpCommand("user", "delete", user.ID)
	if err != nil {
		return err
	}

	return nil
}

// Suspend suspends the current user by executing the appropriate CLI command.
// It sends a request to suspend the user identified by their ID and returns
// the updated user object if successful.
//
// Returns:
//   - A pointer to the updated User object with the suspension applied.
//   - An error if the suspension process fails or if the response cannot be unmarshaled.
func (user *User) Suspend() (*User, error) {
	// Execute the command to suspend a user by ID
	output, err := user.cli.ExecuteOpCommand("user", "suspend", user.ID)
	if err != nil {
		return nil, err
	}

	var updatedUser User
	err = json.Unmarshal([]byte(output), &updatedUser)
	if err != nil {
		return nil, err
	}

	updatedUser.cli = user.cli

	return &updatedUser, nil
}

// Reactivate reactivates a deactivated user in the system.
//
// This method sends a command to the 1Password CLI to reactivate the user
// associated with the current User instance. The reactivation is performed
// using the user's unique ID.
//
// Returns:
//   - nil if the reactivation is successful.
//   - An error if the reactivation command fails or encounters an issue.
//
// Usage:
//
//	err := user.Reactivate()
//	if err != nil {
//	    log.Fatalf("Failed to reactivate user: %v", err)
//	}
//
// Note:
//
//	Ensure that the 1Password CLI is properly configured and authenticated
//	before calling this method, as it relies on the CLI to execute the command.
func (user *User) Reactivate() error {
	// Execute the command to reactivate a user by ID
	_, err := user.cli.ExecuteOpCommand("user", "reactivate", user.ID)
	if err != nil {
		return err
	}

	return nil
}

// SetTravelMode enables or disables travel mode for a user.
// It uses the "op user edit" command to update the travel mode setting.
//
// Parameters:
// - enabled: A boolean indicating whether to enable or disable travel mode.
//
// Returns:
// - An error if the command fails.
func (user *User) SetTravelMode(enabled bool) error {
	// Execute the command to set travel mode for a user by ID
	_, err := user.cli.ExecuteOpCommand("user", "edit", user.ID, fmt.Sprintf("--travel-mode=%t", enabled))
	if err != nil {
		return err
	}

	return nil
}

// SetName updates the name of the user by executing a command with the user's ID.
// It uses the 1Password CLI to perform the operation.
//
// Parameters:
//   - name: The new name to set for the user.
//
// Returns:
//   - error: An error if the command execution fails, otherwise nil.
func (user *User) SetName(name string) error {
	// Execute the command to set the name for a user by ID
	_, err := user.cli.ExecuteOpCommand("user", "edit", user.ID, fmt.Sprintf("--name=%s", name))
	if err != nil {
		return err
	}

	return nil
}
