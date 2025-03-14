# onepassword-cli-go

`onepassword-cli-go` is a Go library for interacting with the 1Password CLI. It provides a set of utilities to manage accounts, items, and vaults programmatically, enabling seamless integration with 1Password in your Go applications.

## Features

- **Account Management**:
  - Retrieve account details by UUID, email, or URL.
  - Check session validity and expiration.
  - Sign in to accounts with passwordless or password-based authentication.

- **Item Management**:
  - Define and manage 1Password items, including fields, sections, and URLs.
  - Support for various item categories (e.g., Login, Password, Secure Note).

- **Vault Management**:
  - Represent and interact with 1Password vaults.

- **CLI Integration**:
  - Execute 1Password CLI commands with support for interactive and non-interactive modes.
  - Verify the integrity of the 1Password CLI executable.

## Installation

To use this library, add it to your Go project:

```bash
go get github.com/sthayduk/onepassword-cli-go
```

Ensure you have the 1Password CLI (`op`) installed and available in your system's PATH.

## Usage

### Initialize the CLI

```go
package main

import (
	"log"
	"github.com/sthayduk/onepassword-cli-go"
)

func main() {
	opPath, err := onepassword.FindOpExecutable()
	if err != nil {
		log.Fatalf("Failed to find 1Password CLI: %v", err)
	}

	cli := &onepassword.OpCLI{
		path: opPath,
	}

	isAvailable, err := onepassword.TestOpCli(opPath)
	if err != nil || !isAvailable {
		log.Fatalf("1Password CLI is not functional: %v", err)
	}

	log.Println("1Password CLI is ready to use.")
}
```

### Account Management

Retrieve account details:

```go
accounts, err := cli.GetAccountDetails()
if err != nil {
	log.Fatalf("Failed to retrieve accounts: %v", err)
}

for _, account := range accounts {
	log.Printf("Account: %s (%s)", account.Email, account.URL)
}
```

Sign in to an account:

```go
account := &onepassword.Account{
	UserUUID: "your-user-uuid",
	Email:    "your-email@example.com",
}

if err := cli.SignIn(account); err != nil {
	log.Fatalf("Failed to sign in: %v", err)
}

log.Println("Signed in successfully!")
```

### Item Management

Define and interact with items:

```go
item := onepassword.Item{
	Title:    "Example Login",
	Category: onepassword.CategoryLogin,
	Fields: []onepassword.Field{
		{
			Label:   "Username",
			Value:   "example_user",
			Type:    onepassword.FieldTypeString,
			Purpose: onepassword.PurposeUsername,
		},
		{
			Label:   "Password",
			Value:   "example_password",
			Type:    onepassword.FieldTypeConcealed,
			Purpose: onepassword.PurposePassword,
		},
	},
}

log.Printf("Created item: %s", item.Title)
```

## Development

### Project Structure

- `accounts.go`: Handles account-related operations, including sign-in and session management.
- `client.go`: Provides the core CLI integration and command execution logic.
- `items.go`: Defines structures and utilities for managing 1Password items.
- `go.mod`: Specifies module dependencies.

### Dependencies

- `golang.org/x/term`: Used for secure password input.
- `golang.org/x/sys`: Provides system-level utilities (indirect dependency).

### Testing

To test the library, ensure the `op` CLI is installed and functional. Use the provided methods to interact with your 1Password environment.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Acknowledgments

- [1Password CLI](https://developer.1password.com/docs/cli): The official CLI for 1Password.
- [Go Programming Language](https://golang.org): The language powering this library.