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

	// Sign in to 1Password
	ctx := context.Background()
	account, err := cli.GetAccountDetailsByEmail("stefan.hayduk@gmail.com")
	if err != nil {
		log.Fatalf("Failed to retrieve account details: %v", err)
	}

	if err := cli.SignIn(ctx, account); err != nil {
		log.Fatalf("Failed to sign in: %v", err)
	}

	// Example: Retrieve all items
	items, err := cli.GetItems()
	if err != nil {
		log.Fatalf("Failed to retrieve items: %v", err)
	}

	fmt.Println("Items:")
	for _, item := range *items {
		fmt.Printf("ID: %s, Title: %s, Category: %s\n", item.ID, item.Title, item.Category)
	}

	// Example: Create Item from Template
	templates, err := cli.GetItemTemplates()
	if err != nil {
		log.Fatalf("Failed to retrieve templates: %v", err)
	}
	fmt.Println("Templates:")
	for _, template := range *templates {
		fmt.Printf("ID: %s, Title: %s\n", template.UUID, template.Name)
	}

	itemTemplate, err := cli.GetItemTemplateByName("Login")
	if err != nil {
		log.Fatalf("Failed to create item from template: %v", err)
	}

	itemTemplate.Title = "Example Item"
	itemTemplate.AddUserName("exampleuser")

	if err := cli.CreateItem(itemTemplate, true); err != nil {
		log.Fatalf("Failed to create item from template: %v", err)
	}

	// Example: Add a URL to an item
	item, err := cli.GetItemByName(itemTemplate.Title)
	if err != nil {
		log.Fatalf("Failed to retrieve item: %v", err)
	}

	newURL := onepassword.ItemURL{
		Href:    "https://example.com",
		Label:   "Example URL",
		Primary: false,
	}
	item.AddURL(newURL)

	if err := item.Save(); err != nil {
		log.Fatalf("Failed to save item: %v", err)
	}

	fmt.Println("Added new URL to item.")

	// Example: Add multiple URLs to an item
	urlsToAdd := []onepassword.ItemURL{
		{
			Href:    "https://example1.com",
			Label:   "Example URL 1",
			Primary: false,
		},
		{
			Href:    "https://example2.com",
			Label:   "Example URL 2",
			Primary: false,
		},
	}

	for _, url := range urlsToAdd {
		item.AddURL(url)
	}

	if err := item.Save(); err != nil {
		log.Fatalf("Failed to save item after adding URLs: %v", err)
	}

	fmt.Println("Added multiple URLs to item.")

	// Example: Remove some URLs from an item
	urlsToRemove := []string{"https://example.com"}
	for _, url := range urlsToRemove {
		if err := item.RemoveURLs(url); err != nil {
			log.Fatalf("Failed to remove URL %s: %v", url, err)
		}
	}

	if err := item.Save(); err != nil {
		log.Fatalf("Failed to save item after URL removal: %v", err)
	}

	fmt.Println("Removed selected URLs from item.")

	// Example: Delete an item
	if err := item.Delete(); err != nil {
		log.Fatalf("Failed to delete item: %v", err)
	}

	fmt.Println("Deleted item.")
}
