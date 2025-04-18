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
	account, err := cli.GetAccountDetailsByEmail("stefan.hayduk@itdesign.at")
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

	if _, err := cli.CreateItem(itemTemplate, true); err != nil {
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
		if err := item.DeleteURLs(url); err != nil {
			log.Fatalf("Failed to remove URL %s: %v", url, err)
		}
	}

	if err := item.Save(); err != nil {
		log.Fatalf("Failed to save item after URL removal: %v", err)
	}

	fmt.Println("Removed selected URLs from item.")

	// Example: Add a new section and fields to an item
	newSection := onepassword.Section{
		ID:    "example-section",
		Label: "Example Section",
	}

	if err := item.AddSection(newSection); err != nil {
		log.Fatalf("Failed to add section: %v", err)
	}

	newField := item.NewField("Example Field", "Example Value", onepassword.FieldTypeConcealed)

	item.AddField(newField)

	if err := item.AddFieldToSection(newSection, newField); err != nil {
		log.Fatalf("Failed to add field to section: %v", err)
	}

	if err := item.Save(); err != nil {
		log.Fatalf("Failed to save item after adding section and field: %v", err)
	}

	fmt.Println("Added new section and field to item.")

	// Example: Remove a section and its fields from an item
	if err := item.DeleteSection(newSection); err != nil {
		log.Fatalf("Failed to delete section: %v", err)
	}

	if err := item.Save(); err != nil {
		log.Fatalf("Failed to save item after deleting section: %v", err)
	}

	fmt.Println("Deleted section and its fields from item.")

	// Add Tag to item
	item.AddTag("example-tag")
	item.AddTag("example-tag-2/credit-card")
	if err := item.Save(); err != nil {
		log.Fatalf("Failed to save item after adding tags: %v", err)
	}
	fmt.Println("Added tags to item.")

	// Example: Remove a tag from an item
	if err := item.DeleteTag("example-tag"); err != nil {
		log.Fatalf("Failed to remove tag: %v", err)
	}
	if err := item.Save(); err != nil {
		log.Fatalf("Failed to save item after removing tag: %v", err)
	}
	fmt.Println("Removed tag from item.")

	// Example: Delete an item
	if err := item.Delete(); err != nil {
		log.Fatalf("Failed to delete item: %v", err)
	}

	fmt.Println("Deleted item.")
}
