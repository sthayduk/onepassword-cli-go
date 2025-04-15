package main

import (
	"context"
	"log"

	"github.com/sthayduk/onepassword-cli-go"
)

func main() {
	// Initialize the 1Password CLI client
	cli := onepassword.NewOpCLI()

	// Sign in to 1Password
	ctx := context.Background()
	account, err := cli.GetAccountDetailsByEmail("stefan.hayduk@itdesign.at")
	if err != nil {
		log.Fatalf("Failed to retrieve account details: %v", err)
	}

	if err := cli.SignIn(ctx, account); err != nil {
		log.Fatalf("Failed to sign in: %v", err)
	}

	// Get the list of groups
	groups, err := cli.GetGroups()
	if err != nil {
		log.Fatalf("Failed to retrieve groups: %v", err)
	}

	// Print the list of groups
	for _, group := range groups {
		log.Printf("Group ID: %s, Name: %s", group.ID, group.Name)
	}

	group, err := cli.GetGroupByName("grp1P-RiskExperts-EMS")
	if err != nil {
		log.Fatalf("Failed to retrieve group: %v", err)
	}
	log.Printf("Retrieved Group ID: %s, Name: %s", group.ID, group.Name)

	members, err := group.ListMembers()
	if err != nil {
		log.Fatalf("Failed to retrieve group members: %v", err)
	}
	for _, member := range members {
		log.Printf("Member ID: %s, Name: %s", member.ID, member.Name)
	}

	// Add new group to 1Password
	newGroup, err := cli.CreateGroup("New Group", "This is a new group")
	if err != nil {
		log.Fatalf("Failed to create new group: %v", err)
	}
	log.Printf("Created new group with ID: %s, Name: %s", newGroup.ID, newGroup.Name)

	// Add a member to the new group
	memberID := "user.email@example.com" // Replace with the actual member ID
	user, err := cli.GetUserByEmail(memberID)
	if err != nil {
		log.Fatalf("Failed to retrieve user: %v", err)
	}

	if err := newGroup.AddMember(*user); err != nil {
		log.Fatalf("Failed to add member to group: %v", err)
	}

	if err := newGroup.AddManager(*user); err != nil {
		log.Fatalf("Failed to add manager to group: %v", err)
	}

	// Remove a member from the new group
	if err := newGroup.RemoveMember(*user); err != nil {
		log.Fatalf("Failed to remove member from group: %v", err)
	}

	// Delete the new group
	if err := newGroup.Delete(); err != nil {
		log.Fatalf("Failed to delete group: %v", err)
	}

}
