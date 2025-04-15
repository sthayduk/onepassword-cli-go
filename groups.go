package onepassword

import (
	"encoding/json"
	"time"
)

type Group struct {
	cli *OpCLI `json:"-"` // Reference to the OpCLI instance for update operations

	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Permissions []string  `json:"permissions,omitempty"`
	Type        string    `json:"type"`
}

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

func (cli *OpCLI) GetGroupByName(name string) (*Group, error) {
	return cli.getGroup(name)
}

func (cli *OpCLI) GetGroupByID(id string) (*Group, error) {
	return cli.getGroup(id)
}

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

func (group *Group) Delete() error {
	// Execute the command to delete a group
	_, err := group.cli.ExecuteOpCommand("group", "delete", group.ID)
	if err != nil {
		return err
	}

	return nil
}
func (group *Group) SetName(name string) error {
	// Execute the command to set the group name
	_, err := group.cli.ExecuteOpCommand("group", "edit", group.ID, "--name", name)
	if err != nil {
		return err
	}

	return nil
}

func (group *Group) SetDescription(description string) error {
	// Execute the command to set the group description
	_, err := group.cli.ExecuteOpCommand("group", "edit", group.ID, "--description", description)
	if err != nil {
		return err
	}

	return nil
}

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
