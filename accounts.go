package onepassword

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

// ErrMultipleAccounts is an error that indicates multiple accounts were found
// when only one was expected. This error can be used to signal ambiguity in
// account selection or retrieval operations.
var ErrMultipleAccounts = errors.New("multiple accounts found")

// Account represents a 1Password account with associated details.
// It includes the account's URL, the email address of the user,
// the unique identifier for the user (UserUUID), and the unique
// identifier for the account (AccountUUID).
type Account struct {
	URL                  string `json:"url"`
	Email                string `json:"email"`
	UserUUID             string `json:"user_uuid"`
	AccountUUID          string `json:"account_uuid"`
	signInTime           time.Time
	signInExpireDuration time.Duration
	sessionToken         string
}

// setSignInTime updates the Account's signInTime field to the current time.
// This method is used to record the time when the account signs in.
func (a *Account) setSignInTime() {
	a.signInTime = time.Now()
}

// setSignInExpireDuration sets the duration after which the account's sign-in session will expire.
// This method updates the signInExpireDuration field of the Account struct.
//
// Parameters:
//   - duration: The time.Duration value representing the expiration duration.
func (a *Account) setSignInExpireDuration(duration time.Duration) {
	a.signInExpireDuration = duration
}

// setSessionToken sets the session token for the account.
// This method updates the sessionToken field of the Account struct.
//
// Parameters:
//   - token: The session token to be associated with the account.
func (a *Account) setSessionToken(token string) {
	a.sessionToken = token
}

// SetSignInInfo updates the account's sign-in information by setting the
// sign-in time, session token, and session expiration duration.
//
// Parameters:
//   - token: The session token to be associated with the account.
//
// Notes:
//   - Session tokens expire after 30 minutes of inactivity. This method
//     sets the expiration duration to 29 minutes to provide a buffer.
func (a *Account) SetSignInInfo(token string) {
	a.setSignInTime()
	a.setSessionToken(token)

	// Session tokens expire after 30 minutes of inactivity, after which you’ll need to sign in again.
	a.setSignInExpireDuration(1740 * time.Second) // 29 minutes to have some buffer
}

// IsSessionExpired checks if the session associated with the account has expired.
// It compares the time elapsed since the account's sign-in time with the session's
// expiration duration. Returns true if the session has expired, otherwise false.
func (a *Account) IsSessionExpired() bool {
	return time.Since(a.signInTime) > a.signInExpireDuration
}

// IsSessionValid checks if the session associated with the account is valid.
// It returns true if the session is not expired, otherwise false.
func (a *Account) IsSessionValid() bool {
	return !a.IsSessionExpired()
}

// GetAccountDetails retrieves the details of all 1Password accounts configured
// in the CLI. It executes the "op account list" command, parses the result,
// and returns a slice of Account objects.
//
// Returns:
//   - ([]Account): A slice of Account objects representing the 1Password accounts.
//   - (error): An error if the command execution or JSON parsing fails, or if no
//     accounts are found.
//
// Errors:
//   - Returns an error if the "op account list" command fails to execute.
//   - Returns an error if the JSON output cannot be parsed into Account objects.
//   - Returns an error if no accounts are found.
func (cli *OpCLI) GetAccountDetails() ([]Account, error) {
	slog.Debug("retrieving 1Password account details")

	output, err := exec.Command(cli.Path, "account", "list", "--format=json").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %v", err)
	}

	var accounts []Account
	if err := json.Unmarshal(output, &accounts); err != nil {
		return nil, fmt.Errorf("failed to parse account list: %v", err)
	}

	if len(accounts) == 0 {
		slog.Error("no 1Password accounts found")
		return nil, fmt.Errorf("no accounts found")
	}

	return accounts, nil
}

// GetAccountDetailsByUUID retrieves the details of a 1Password account by its UUID.
// It searches through the list of accounts obtained from the GetAccountDetails method.
// If an account with the specified UUID is found, it returns the account details.
// Otherwise, it returns an error indicating that the account was not found.
//
// Parameters:
//   - accountUUID: The UUID of the account to retrieve.
//
// Returns:
//   - *Account: A pointer to the Account struct containing the account details.
//   - error: An error if the account is not found or if there is an issue retrieving the account details.
func (cli *OpCLI) GetAccountDetailsByUUID(accountUUID string) (*Account, error) {
	slog.Debug("retrieving 1Password account details by UUID", "accountUUID", accountUUID)

	accounts, err := cli.GetAccountDetails()
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if account.UserUUID == accountUUID {
			return &account, nil
		}
	}

	return nil, fmt.Errorf("account with UUID %s not found", accountUUID)
}

// GetAccountDetailsByEmail retrieves the details of a 1Password account
// associated with the specified email address.
//
// This method fetches all available account details using the GetAccountDetails
// method and searches for an account that matches the provided email.
//
// Parameters:
//   - email: The email address of the account to retrieve.
//
// Returns:
//   - A pointer to the Account struct if an account with the specified email is found.
//   - An error if no account with the specified email is found or if there is an issue
//     retrieving account details.
func (cli *OpCLI) GetAccountDetailsByEmail(email string) (*Account, error) {
	slog.Debug("retrieving 1Password account details by email", "email", email)

	accounts, err := cli.GetAccountDetails()
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if account.Email == email {
			return &account, nil
		}
	}

	return nil, fmt.Errorf("account with email %s not found", email)
}

// GetAccountDetailsByURL retrieves the details of a 1Password account that matches the specified URL.
// It searches through all available accounts and returns the account details if a match is found.
//
// Parameters:
//   - url: The URL of the 1Password account to retrieve.
//
// Returns:
//   - *Account: A pointer to the matching Account object if found.
//   - error: An error if no account is found, multiple accounts match the URL, or if there is an issue retrieving account details.
//
// Errors:
//   - Returns an error if no account matches the specified URL.
//   - Returns an error if multiple accounts match the specified URL.
//   - Returns an error if there is an issue retrieving the account details.
func (cli *OpCLI) GetAccountDetailsByURL(url string) (*Account, error) {
	slog.Debug("retrieving 1Password account details by URL", "url", url)

	accounts, err := cli.GetAccountDetails()
	if err != nil {
		return nil, err
	}

	matchingAccounts := []Account{}
	for _, account := range accounts {
		if normalizeURL(account.URL) == normalizeURL(url) {
			matchingAccounts = append(matchingAccounts, account)
		}
	}

	if len(matchingAccounts) > 1 {
		return nil, fmt.Errorf("%w: URL %s", ErrMultipleAccounts, url)
	}

	if len(matchingAccounts) == 1 {
		return &matchingAccounts[0], nil
	}

	return nil, fmt.Errorf("account with URL %s not found", url)
}

// GetAccountDetailsByAccountUUID retrieves the details of a 1Password account
// by its unique account UUID.
//
// This method fetches all available account details using the GetAccountDetails
// method and searches for the account that matches the provided UUID. If a match
// is found, it returns the account details. If no match is found, an error is returned.
//
// Parameters:
//   - accountUUID: A string representing the unique identifier of the account.
//
// Returns:
//   - *Account: A pointer to the Account struct containing the account details, if found.
//   - error: An error if the account with the specified UUID is not found or if there
//     is an issue retrieving the account details.
func (cli *OpCLI) GetAccountDetailsByAccountUUID(accountUUID string) (*Account, error) {
	slog.Debug("retrieving 1Password account details by account UUID", "accountUUID", accountUUID)

	accounts, err := cli.GetAccountDetails()
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if account.AccountUUID == accountUUID {
			return &account, nil
		}
	}

	return nil, fmt.Errorf("account with UUID %s not found", accountUUID)
}

// normalizeURL standardizes URLs by removing protocols and trailing paths.
//
// Parameters:
//   - url: The URL to normalize
//
// Returns:
//   - string: The normalized URL in lowercase
func normalizeURL(url string) string {
	url = strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}
	return strings.ToLower(url)
}
