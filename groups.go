package onepassword

import (
	"encoding/json"
	"time"
)

type Group struct {
	cli *OpCLI `json:"-"` // Reference to the OpCLI instance for update operations

	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	State       string       `json:"state"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Permissions []Permission `json:"permissions,omitempty"`
	Type        string       `json:"type"`
}

// ListGroups retrieves a list of all groups available in the 1Password CLI.
// It executes the "group list" command and parses the output into a slice of Group objects.
//
// Returns:
//   - ([]Group): A slice of Group objects.
//   - (error): An error if the operation fails.
func (cli *OpCLI) ListGroups() ([]Group, error) {

	output, err := cli.ExecuteOpCommand("group", "list")
	if err != nil {
		return nil, err
	}

	var groups []Group
	err = json.Unmarshal([]byte(output), &groups)
	if err != nil {
		return nil, err
	}

	// Set the cli reference for each group
	for i := range groups {
		groups[i].cli = cli
	}

	return groups, nil
}

// getGroup retrieves a specific group by its ID or name.
// It executes the "group get" command and parses the output into a Group object.
//
// Parameters:
//   - groupID (string): The ID or name of the group to retrieve.
//
// Returns:
//   - (*Group): A pointer to the Group object.
//   - (error): An error if the operation fails.
func (cli *OpCLI) getGroup(groupID string) (*Group, error) {
	// Execute the command to get a group by ID
	output, err := cli.ExecuteOpCommand("group", "get ", groupID)
	if err != nil {
		return nil, err
	}

	var group Group
	err = json.Unmarshal([]byte(output), &group)
	if err != nil {
		return nil, err
	}

	group.cli = cli

	return &group, nil
}

// GetGroupByName retrieves a group by its name.
// It internally calls getGroup with the group name.
//
// Parameters:
//   - name (string): The name of the group to retrieve.
//
// Returns:
//   - (*Group): A pointer to the Group object.
//   - (error): An error if the operation fails.
func (cli *OpCLI) GetGroupByName(name string) (*Group, error) {
	return cli.getGroup(name)
}

// GetGroupByID retrieves a group by its ID.
// It internally calls getGroup with the group ID.
//
// Parameters:
//   - id (string): The ID of the group to retrieve.
//
// Returns:
//   - (*Group): A pointer to the Group object.
//   - (error): An error if the operation fails.
func (cli *OpCLI) GetGroupByID(id string) (*Group, error) {
	return cli.getGroup(id)
}

// CreateGroup creates a new group with the specified name and description.
// It executes the "group create" command and parses the output into a Group object.
//
// Parameters:
//   - name (string): The name of the group to create.
//   - description (string): The description of the group.
//
// Returns:
//   - (*Group): A pointer to the newly created Group object.
//   - (error): An error if the operation fails.
func (cli *OpCLI) CreateGroup(name string, description string) (*Group, error) {
	// Execute the command to create a group
	output, err := cli.ExecuteOpCommand("group", "create", name, "--description", description)
	if err != nil {
		return nil, err
	}

	var group Group
	err = json.Unmarshal([]byte(output), &group)
	if err != nil {
		return nil, err
	}

	group.cli = cli

	return &group, nil
}

// Delete removes the group from the 1Password CLI.
// It executes the "group delete" command using the group's ID.
//
// Returns:
//   - (error): An error if the operation fails.
func (group *Group) Delete() error {
	// Execute the command to delete a group
	_, err := group.cli.ExecuteOpCommand("group", "delete", group.ID)
	if err != nil {
		return err
	}

	return nil
}

// SetName updates the name of the group.
// It executes the "group edit" command with the new name.
//
// Parameters:
//   - name (string): The new name for the group.
//
// Returns:
//   - (error): An error if the operation fails.
func (group *Group) SetName(name string) error {
	// Execute the command to set the group name
	_, err := group.cli.ExecuteOpCommand("group", "edit", group.ID, "--name", name)
	if err != nil {
		return err
	}

	return nil
}

// SetDescription updates the description of the group.
// It executes the "group edit" command with the new description.
//
// Parameters:
//   - description (string): The new description for the group.
//
// Returns:
//   - (error): An error if the operation fails.
func (group *Group) SetDescription(description string) error {
	// Execute the command to set the group description
	_, err := group.cli.ExecuteOpCommand("group", "edit", group.ID, "--description", description)
	if err != nil {
		return err
	}

	return nil
}

// ListMembers retrieves a list of all users who are members of the group.
// It executes the "group user list" command and parses the output into a slice of User objects.
//
// Returns:
//   - ([]User): A slice of User objects.
//   - (error): An error if the operation fails.
func (group *Group) ListMembers() ([]User, error) {
	// Execute the command to list group members
	output, err := group.cli.ExecuteOpCommand("group", "user", "list", group.ID)
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
		users[i].cli = group.cli
	}

	return users, nil
}

// AddMember adds a user to the group with the default role of "member".
// It executes the "group user grant" command with the user's ID and the group's ID.
//
// Parameters:
//   - user (User): The user to add to the group.
//
// Returns:
//   - (error): An error if the operation fails.
func (group *Group) AddMember(user User) error {
	// Execute the command to add a user to the group
	_, err := group.cli.ExecuteOpCommand("group", "user", "grant",
		"--group", group.ID,
		"--user", user.ID,
		"--role", "member")
	// --role is optional, default is "member"
	if err != nil {
		return err
	}

	return nil
}

// RemoveMember removes a user from the group.
// It executes the "group user revoke" command with the user's ID and the group's ID.
//
// Parameters:
//   - user (User): The user to remove from the group.
//
// Returns:
//   - (error): An error if the operation fails.
func (group *Group) RemoveMember(user User) error {
	// Execute the command to remove a user from the group
	_, err := group.cli.ExecuteOpCommand("group", "user", "revoke",
		"--group", group.ID,
		"--user", user.ID)

	if err != nil {
		return err
	}

	return nil
}

// AddManager adds a user to the group with the role of "manager".
// It executes the "group user grant" command with the user's ID and the group's ID.
//
// Parameters:
//   - user (User): The user to add as a manager to the group.
//
// Returns:
//   - (error): An error if the operation fails.
func (group *Group) AddManager(user User) error {
	// Execute the command to add a manager to the group
	_, err := group.cli.ExecuteOpCommand("group", "user", "grant",
		"--group", group.ID,
		"--user", user.ID,
		"--role", "manager")
	// --role is optional, default is "member"
	if err != nil {
		return err
	}

	return nil
}

// RemoveManager removes a user from the group who has the role of "manager".
// It executes the "group user revoke" command with the user's ID and the group's ID.
//
// Parameters:
//   - user (User): The user to remove as a manager from the group.
//
// Returns:
//   - (error): An error if the operation fails.
func (group *Group) RemoveManager(user User) error {
	// Execute the command to remove a manager from the group
	_, err := group.cli.ExecuteOpCommand("group", "user", "revoke",
		"--group", group.ID,
		"--user", user.ID)

	if err != nil {
		return err
	}

	return nil
}
