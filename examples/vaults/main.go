package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sthayduk/onepassword-cli-go"
)

func main() {
	// Initialize the 1Password CLI client
	cli := onepassword.NewOpCLI()

	// Create a context for the sign-in operation
	ctx := context.Background()

	// Sign in to 1Password
	account, err := cli.GetAccountDetailsByEmail("stefan.hayduk@itdesign.at")
	if err != nil {
		log.Fatalf("Failed to retrieve account details: %v", err)
	}

	if err := cli.SignIn(ctx, account); err != nil {
		log.Fatalf("Failed to sign in: %v", err)
	}

	// Example: Get all vault details
	vaults, err := cli.GetVaultDetails()
	if err != nil {
		log.Fatalf("Failed to retrieve vaults: %v", err)
	}

	fmt.Println("Vaults:")
	for _, vault := range *vaults {
		fmt.Printf("ID: %s, Name: %s, Items: %d\n", vault.ID, vault.Name, vault.Items)
	}

	// Example: Get details of a specific vault by ID
	vaultID := "bbq6cjaznuofmtvfv5nejy36eq" // Replace with a valid vault ID
	vault, err := cli.GetVaultDetailsByID(vaultID)
	if err != nil {
		log.Fatalf("Failed to retrieve vault details for ID %s: %v", vaultID, err)
	}

	fmt.Printf("\nDetails of Vault ID %s:\n", vaultID)
	fmt.Printf("Name: %s\nContent Version: %d\nCreated At: %s\nUpdated At: %s\nItems: %d\nDescription: %s\nType: %s\n",
		vault.Name, vault.ContentVersion, vault.CreatedAt, vault.UpdatedAt, vault.Items, vault.Description, vault.Type)

	// Example: Create a new vault
	newVault, err := cli.CreateVault("New Vault", "This is a vault description.", onepassword.IconApplication, true)
	if err != nil {
		log.Fatalf("Failed to create new vault: %v", err)
	}
	fmt.Printf("\nCreated new vault: ID: %s, Name: %s\n", newVault.ID, newVault.Name)

	// Example: Update an existing vault
	if err := newVault.SetIcon(onepassword.IconAirplane); err != nil {
		log.Fatalf("Failed to set icon for vault: %v", err)
	}

	// Example: Set Permissions for the vault
	group, err := cli.GetGroupByName("GroupName") // Replace with a valid group name
	if err != nil {
		log.Fatalf("Failed to retrieve group details: %v", err)
	}

	if err := newVault.GrantGroupPermission(*group, onepassword.PermissionCopyAndShareItems); err != nil {
		log.Fatalf("Failed to grant group permission: %v", err)
	}
	fmt.Printf("Granted group permission for group: %s\n", group.Name)

	// Example: Set Permissions for the vault
	user, err := cli.GetUserByEmail("user.mail@example.com") // Replace with a valid user email
	if err != nil {
		log.Fatalf("Failed to retrieve user details: %v", err)
	}

	if err := newVault.GrantUserPermission(*user, onepassword.PermissionMoveItems); err != nil {
		log.Fatalf("Failed to grant user permission: %v", err)
	}
	fmt.Printf("Granted user permission for user: %s\n", user.Email)

	// Example: Delete a vault
	if err := newVault.Delete(); err != nil {
		log.Fatalf("Failed to delete vault: %v", err)
	}
	fmt.Printf("Deleted vault with ID: %s\n", newVault.ID)

}
