package main

import (
	"log"

	"github.com/sthayduk/onepassword-cli-go"
)

func main() {

	cli := onepassword.NewOpCLI()
	err := cli.SignInWithServiceAccount("your-service-account-token")
	if err != nil {
		log.Fatalf("Failed to sign in: %v", err)
	}

	items, err := cli.GetItems()
	if err != nil {
		log.Fatalf("Failed to get items: %v", err)
	}

	for _, item := range *items {
		log.Printf("Item: %s", item.Title)
	}

	log.Println("Finished processing items.")

}
