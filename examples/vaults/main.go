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
	for _, vault := range vaults {
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
}
