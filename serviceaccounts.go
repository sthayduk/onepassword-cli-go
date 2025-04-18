package onepassword

import (
	"encoding/json"
	"errors"
	"os"
)

// ServiceAccountRateLimit represents the rate limit information for a service account action.
// It includes the type of rate limit, the action being limited, the maximum allowed requests (Limit),
// the number of requests used (Used), the number of requests remaining (Remaining),
// and the time when the rate limit resets (Reset, as a Unix timestamp).
type ServiceAccountRateLimit struct {
	Type      string `json:"type"`
	Action    string `json:"action"`
	Limit     int    `json:"limit"`
	Used      int    `json:"used"`
	Remaining int    `json:"remaining"`
	Reset     int64  `json:"reset"` // Time in seconds until the rate limit resets
}

// GetServiceAccountRateLimits retrieves the current rate limit information for the authenticated service account.
//
// This method checks if the OpCLI instance is authenticated as a service account. If not, it returns an error.
// It then executes the "service-account rate-limit" command using the 1Password CLI and parses the resulting JSON output
// into a ServiceAccountRateLimit struct. If any step fails (authentication check, command execution, or JSON unmarshalling),
// an appropriate error is returned.
//
// Returns:
//   - []ServiceAccountRateLimit: Slice containing the current rate limit details for the service account.
//   - error: Non-nil if the operation fails due to authentication, command execution, or parsing errors.
//
// Example usage:
//
//	rateLimits, err := cli.GetServiceAccountRateLimits()
//	if err != nil {
//	    log.Fatalf("Failed to get rate limit: %v", err)
//	}
//	fmt.Printf("Remaining requests: %d\n", rateLimits[0].Remaining)
func (cli *OpCLI) GetServiceAccountRateLimits() ([]ServiceAccountRateLimit, error) {

	if !cli.isServiceAccount {
		return []ServiceAccountRateLimit{}, errors.New("not authenticated as a service account")
	}

	output, err := cli.ExecuteOpCommand("service-account", "rate-limit")
	if err != nil {
		return []ServiceAccountRateLimit{}, err
	}

	var rateLimits []ServiceAccountRateLimit
	err = json.Unmarshal(output, &rateLimits)
	if err != nil {
		return []ServiceAccountRateLimit{}, err
	}

	return rateLimits, nil
}

// SignInWithServiceAccount authenticates the OpCLI instance using a 1Password service account access token.
//
// This method sets the provided access token as the current authentication token for the CLI instance,
// marks the instance as authenticated via a service account, and sets the "OP_SERVICE_ACCOUNT_TOKEN"
// environment variable for downstream processes. It then retrieves the current user's details using the
// GetMe method and updates the OpCLI's Account field with the user's UUID and email.
//
// Parameters:
//   - accesstoken: A string representing the 1Password service account access token.
//
// Returns:
//   - error: Returns an error if retrieving the user details fails; otherwise, returns nil.
//
// Side Effects:
//   - Modifies the OpCLI instance's accesstoken and isServiceAccount fields.
//   - Sets the "OP_SERVICE_ACCOUNT_TOKEN" environment variable.
//   - Updates the OpCLI's Account field with the authenticated user's details.
//
// Example usage:
//
//	err := cli.SignInWithServiceAccount("your-access-token")
//	if err != nil {
//	    log.Fatalf("Failed to sign in: %v", err)
//	}
func (cli *OpCLI) SignInWithServiceAccount(accesstoken string) error {
	cli.accesstoken = accesstoken
	cli.isServiceAccount = true

	os.Setenv("OP_SERVICE_ACCOUNT_TOKEN", cli.accesstoken)

	user, err := cli.GetMe()
	if err != nil {
		return err
	}

	// Set the user details in the OpCLI instance
	// This is necessary for the CLI to function properly with the service account
	// and the other signin methods return a account object with the required information
	// but this method does return a Userobject, so we need to set it manually
	cli.Account = &Account{
		UserUUID: user.ID,
		Email:    user.Email,
	}

	return nil
}
