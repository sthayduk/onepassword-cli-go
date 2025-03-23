package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sthayduk/onepassword-cli-go"
)

func main() {
	// Setup logging
	log := log.New(os.Stdout, "[1Password] ", log.LstdFlags)

	// Create context with cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Handle graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Println("Received shutdown signal")
		cancel()
	}()

	if err := signin(ctx, log); err != nil {
		log.Printf("Error during signin: %v\n", err)
		os.Exit(1)
	}
}

func signin(ctx context.Context, log *log.Logger) error {
	clt := onepassword.NewOpCLI()

	// Get email from environment variable or use default
	email := getEnv("OP_EMAIL", "stefan.hayduk@gmail.com")
	log.Printf("Attempting to sign in with email: %s\n", email)

	if err := onepassword.TestOpCli(clt.Path); err != nil {
		return fmt.Errorf("CLI test failed: %w", err)
	}

	if err := onepassword.VerifyOpExecutable(clt.Path); err != nil {
		return fmt.Errorf("CLI verification failed: %w", err)
	}

	account, err := clt.GetAccountDetailsByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to get account details: %w", err)
	}

	if err := clt.SignIn(ctx, account); err != nil {
		return fmt.Errorf("signin failed: %w", err)
	}

	log.Println("Signed in successfully")
	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
