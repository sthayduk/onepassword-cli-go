# onepassword-cli-go

`onepassword-cli-go` is a Go library for interacting with the 1Password CLI. It provides a set of utilities to manage accounts, items, vaults, groups, and permissions programmatically, enabling seamless integration with 1Password in your Go applications.

## Features

- **Account Management**:
  - Retrieve account details by UUID, email, or URL.
  - Check session validity and expiration.
  - Sign in to accounts with passwordless or password-based authentication.
  - Sign in with service account accesstoken

- **Item Management**:
  - Define and manage 1Password items, including fields, sections, and URLs.
  - Support for various item categories (e.g., Login, Password, Secure Note, Identity).
  - Add and delete sections within items, ensuring unique section IDs.
  - Add and delete fields within specific sections, maintaining consistent state.
  - Add and remove URLs associated with items.
  - Save and delete items programmatically.
  - Add tags to items for better organization.

- **Vault Management**:
  - Represent and interact with 1Password vaults.
  - Retrieve vault details by ID or name.
  - Validate vault IDs and update vault icons.
  - Create, delete, and update vaults.

- **Group Management**:
  - List, create, and delete groups.
  - Add and remove members or managers from groups.
  - Update group names and descriptions.

- **Permission Management**:
  - Define and resolve granular permissions for items and vaults.
  - Manage dependencies between permissions.

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

Define and create an item:

```go
item := onepassword.Item{
    Title:    "Example Login",
    Category: onepassword.CategoryLogin,
    Vault: onepassword.Vault{
        ID:   "vault-id", // Replace with a valid vault ID
        Name: "Personal",
    },
    Fields: []onepassword.Field{
        {
            Label:   "Username",
            Value:   "example_user",
            Type:    onepassword.FieldTypeString,
            Purpose: onepassword.FieldPurposeUsername,
        },
        {
            Label:   "Password",
            Value:   "example_password",
            Type:    onepassword.FieldTypeConcealed,
            Purpose: onepassword.FieldPurposePassword,
        },
    },
}

createdItem, err := cli.CreateItem(&item, false) // Set to true to generate a password
if err != nil {
    log.Fatalf("Failed to create item: %v", err)
}

log.Printf("Created item: %s (ID: %s)", createdItem.Title, createdItem.ID)
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

#### Add and Remove URLs

Add a URL to an item:

```go
newURL := onepassword.ItemURL{
    Href:    "https://example.com",
    Label:   "Example URL",
    Primary: true,
}
item.AddURL(newURL)

if err := item.Save(); err != nil {
    log.Fatalf("Failed to save item: %v", err)
}

log.Println("Added new URL to item.")
```

Remove a URL from an item:

```go
err := item.DeleteURLs("https://example.com")
if err != nil {
    log.Fatalf("Failed to remove URL: %v", err)
}

if err := item.Save(); err != nil {
    log.Fatalf("Failed to save item after URL removal: %v", err)
}

log.Println("Removed URL from item.")
```

### Vault Management

Retrieve vault details:

```go
vaults, err := cli.GetVaultDetails()
if err != nil {
    log.Fatalf("Failed to retrieve vaults: %v", err)
}

for _, vault := range *vaults {
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

### Group Management

List all groups:

```go
groups, err := cli.GetGroups()
if err != nil {
    log.Fatalf("Failed to list groups: %v", err)
}

for _, group := range groups {
    log.Printf("Group: %s (%s)", group.Name, group.ID)
}
```

Create a new group:

```go
group, err := cli.CreateGroup("Example Group", "This is an example group.")
if err != nil {
    log.Fatalf("Failed to create group: %v", err)
}

log.Printf("Created group: %s (%s)", group.Name, group.ID)
```

## Development

### Project Structure

- `accounts.go`: Handles account-related operations, including sign-in and session management.
- `client.go`: Provides the core CLI integration and command execution logic.
- `items.go`: Defines structures and utilities for managing 1Password items.
- `vaults.go`: Contains functions for vault-related operations.
- `groups.go`: Manages groups and their members.
- `permissions.go`: Handles permission definitions and dependencies.
- `examples/`: Contains example programs demonstrating library usage.
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
