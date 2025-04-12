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
  - Add and delete sections within items, ensuring unique section IDs.
  - Add and delete fields within specific sections, maintaining consistent state.

- **Vault Management**:
  - Represent and interact with 1Password vaults.
  - Retrieve vault details by ID or name.
  - Validate vault IDs and update vault icons.

- **CLI Integration**:
  - Execute 1Password CLI commands with support for interactive and non-interactive modes.
  - Verify the integrity of the 1Password CLI executable.
  - Centralized command execution with automatic account flag inclusion.

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
    cli := onepassword.NewOpCLI()
    
    if err := onepassword.TestOpCli(cli.Path); err != nil {
        log.Fatalf("1Password CLI is not functional: %v", err)
    }
    
    if err := onepassword.VerifyOpExecutable(cli.Path); err != nil {
        log.Fatalf("CLI verification failed: %v", err)
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
ctx := context.Background()
account, err := cli.GetAccountDetailsByEmail("your-email@example.com")
if err != nil {
    log.Fatalf("Failed to get account details: %v", err)
}

if err := cli.SignIn(ctx, account); err != nil {
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
    Vault: onepassword.Vault{
        ID:   "vault-id",
        Name: "Personal",
    },
    Fields: []onepassword.Field{
        {
            ID:      "username",
            Label:   "Username",
            Value:   "example_user",
            Type:    onepassword.FieldTypeString,
            Purpose: onepassword.PurposeUsername,
        },
        {
            ID:      "password",
            Label:   "Password",
            Value:   "example_password",
            Type:    onepassword.FieldTypeConcealed,
            Purpose: onepassword.PurposePassword,
        },
    },
}

log.Printf("Created item: %s", item.Title)
```

#### Add a Section to an Item

```go
section := onepassword.Section{
    ID:    "section-id",
    Label: "Example Section",
}

err := item.AddSection(section)
if err != nil {
    log.Fatalf("Failed to add section: %v", err)
}

log.Println("Section added successfully!")
```

#### Delete a Section from an Item

```go
err := item.DeleteSection(section)
if err != nil {
    log.Fatalf("Failed to delete section: %v", err)
}

log.Println("Section deleted successfully!")
```

#### Add a Field to a Section

```go
field := onepassword.Field{
    ID:      "field-id",
    Label:   "Example Field",
    Value:   "example_value",
    Type:    onepassword.FieldTypeString,
    Purpose: onepassword.PurposeNotes,
}

err := item.AddFieldToSection(section, field)
if err != nil {
    log.Fatalf("Failed to add field: %v", err)
}

log.Println("Field added successfully!")
```

#### Delete a Field from a Section

```go
err := item.DeleteFieldFromSection(section, field)
if err != nil {
    log.Fatalf("Failed to delete field: %v", err)
}

log.Println("Field deleted successfully!")
```

### Vault Management

Retrieve vault details:

```go
vaults, err := cli.GetVaultDetails()
if err != nil {
    log.Fatalf("Failed to retrieve vaults: %v", err)
}

for _, vault := range vaults {
    log.Printf("Vault: %s (%s)", vault.Name, vault.ID)
}
```

Retrieve a specific vault by ID:

```go
vaultID := "your-vault-id"
vault, err := cli.GetVaultDetailsByID(vaultID)
if err != nil {
    log.Fatalf("Failed to retrieve vault details: %v", err)
}

log.Printf("Vault Name: %s, Items: %d", vault.Name, vault.Items)
```

Update a vault icon:

```go
err := cli.UpdateVaultIcon(vaultID, onepassword.IconTreasureChest)
if err != nil {
    log.Fatalf("Failed to update vault icon: %v", err)
}

log.Println("Vault icon updated successfully!")
```

## Development

### Project Structure

- `accounts.go`: Handles account-related operations, including sign-in and session management.
- `client.go`: Provides the core CLI integration and command execution logic.
- `items.go`: Defines structures and utilities for managing 1Password items.
- `vaults.go`: Contains functions for vault-related operations.
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
