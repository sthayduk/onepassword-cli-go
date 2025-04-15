package onepassword

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"
)

// Vault represents a 1Password vault.
//
// Fields:
// - ID: A unique 26-character alphanumeric identifier for the vault.
// - Name: The name of the vault.
// - ContentVersion: The version of the vault's content, incremented with changes.
// - CreatedAt: The timestamp when the vault was created, in ISO 8601 format.
// - UpdatedAt: The timestamp when the vault was last updated, in ISO 8601 format.
// - Items: The number of items stored in the vault.
// - Description: A brief description of the vault's purpose or contents.
// - AttributeVersion: The version of the vault's attributes.
// - Type: The type of the vault, e.g., USER_CREATED or SYSTEM_GENERATED.
type Vault struct {
	cli *OpCLI `json:"-"` // Reference to the OpCLI instance for update operations

	ID               string `json:"id"`
	Name             string `json:"name"`
	ContentVersion   int    `json:"content_version"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	Items            int    `json:"items"`
	Description      string `json:"description"`
	AttributeVersion int    `json:"attribute_version"`
	Type             string `json:"type"`
}

// VaultIcon represents the valid icon names for a vault.
type VaultIcon string

const (
	IconAirplane         VaultIcon = "airplane"
	IconApplication      VaultIcon = "application"
	IconArtSupplies      VaultIcon = "art-supplies"
	IconBankersBox       VaultIcon = "bankers-box"
	IconBrownBriefcase   VaultIcon = "brown-briefcase"
	IconBrownGate        VaultIcon = "brown-gate"
	IconBuildings        VaultIcon = "buildings"
	IconCabin            VaultIcon = "cabin"
	IconCastle           VaultIcon = "castle"
	IconCircleOfDots     VaultIcon = "circle-of-dots"
	IconCoffee           VaultIcon = "coffee"
	IconColorWheel       VaultIcon = "color-wheel"
	IconCurtainedWindow  VaultIcon = "curtained-window"
	IconDocument         VaultIcon = "document"
	IconDoughnut         VaultIcon = "doughnut"
	IconFence            VaultIcon = "fence"
	IconGalaxy           VaultIcon = "galaxy"
	IconGears            VaultIcon = "gears"
	IconGlobe            VaultIcon = "globe"
	IconGreenBackpack    VaultIcon = "green-backpack"
	IconGreenGem         VaultIcon = "green-gem"
	IconHandshake        VaultIcon = "handshake"
	IconHeartWithMonitor VaultIcon = "heart-with-monitor"
	IconHouse            VaultIcon = "house"
	IconIDCard           VaultIcon = "id-card"
	IconJet              VaultIcon = "jet"
	IconLargeShip        VaultIcon = "large-ship"
	IconLuggage          VaultIcon = "luggage"
	IconPlant            VaultIcon = "plant"
	IconPorthole         VaultIcon = "porthole"
	IconPuzzle           VaultIcon = "puzzle"
	IconRainbow          VaultIcon = "rainbow"
	IconRecord           VaultIcon = "record"
	IconRoundDoor        VaultIcon = "round-door"
	IconSandals          VaultIcon = "sandals"
	IconScales           VaultIcon = "scales"
	IconScrewdriver      VaultIcon = "screwdriver"
	IconShop             VaultIcon = "shop"
	IconTallWindow       VaultIcon = "tall-window"
	IconTreasureChest    VaultIcon = "treasure-chest"
	IconVaultDoor        VaultIcon = "vault-door"
	IconVehicle          VaultIcon = "vehicle"
	IconWallet           VaultIcon = "wallet"
	IconWrench           VaultIcon = "wrench"
)

// GetVaultDetails retrieves a list of all vaults using the 1Password CLI.
//
// This method executes the "vault list" command using the 1Password CLI to fetch details of all vaults.
// It unmarshals the JSON output into a slice of Vault structs and sets the CLI reference for each vault.
//
// Returns:
// - *[]Vault: A pointer to a slice of Vault structs containing details of each vault.
// - error: An error object if the operation fails.
func (cli *OpCLI) GetVaultDetails() (*[]Vault, error) {
	output, err := cli.ExecuteOpCommand("vault", "list")
	if err != nil {
		return nil, err
	}

	var vaults []Vault
	err = json.Unmarshal(output, &vaults)
	if err != nil {
		return nil, err
	}

	// Set the cli reference for each vault
	for i := range vaults {
		vaults[i].cli = cli
	}

	return &vaults, nil
}

// getVaultDetails retrieves the details of a specific vault by its identifier.
//
// This method executes the "vault get" command using the 1Password CLI to fetch details of a specific vault.
// It unmarshals the JSON output into a Vault struct and sets the CLI reference for the vault.
//
// Parameters:
// - identifier: The unique identifier or name of the vault.
//
// Returns:
// - *Vault: A pointer to a Vault struct containing the vault's details.
// - error: An error object if the operation fails.
func (cli *OpCLI) getVaultDetails(identifier string) (*Vault, error) {
	output, err := cli.ExecuteOpCommand("vault", "get", identifier)
	if err != nil {
		return nil, err
	}

	var vault Vault
	err = json.Unmarshal(output, &vault)
	if err != nil {
		return nil, err
	}

	// Set the cli reference for the vault
	vault.cli = cli

	return &vault, nil
}

// GetVaultDetailsByName retrieves the details of a vault by its name.
//
// This method is a wrapper around getVaultDetails, allowing retrieval of vault details using the vault's name.
//
// Parameters:
// - vaultName: The name of the vault.
//
// Returns:
// - *Vault: A pointer to a Vault struct containing the vault's details.
// - error: An error object if the operation fails.
func (cli *OpCLI) GetVaultDetailsByName(vaultName string) (*Vault, error) {
	return cli.getVaultDetails(vaultName)
}

// GetVaultDetailsByID retrieves the details of a vault by its ID.
//
// This method validates the vault ID format and then calls getVaultDetails to fetch the vault details.
//
// Parameters:
// - vaultID: The unique identifier of the vault.
//
// Returns:
// - *Vault: A pointer to a Vault struct containing the vault's details.
// - error: An error object if the operation fails.
func (cli *OpCLI) GetVaultDetailsByID(vaultID string) (*Vault, error) {
	if err := ValidateVaultID(vaultID); err != nil {
		return nil, err
	}

	return cli.getVaultDetails(vaultID)
}

// CreateVault creates a new vault in 1Password.
//
// This method executes the "vault create" command using the 1Password CLI to create a new vault with the specified parameters.
//
// Parameters:
// - name: The name of the new vault.
// - description: A brief description of the vault's purpose or contents.
// - icon: The icon to associate with the vault. Must be a valid VaultIcon.
// - adminAccess: A boolean indicating whether admins are allowed to manage the vault.
//
// Returns:
// - *Vault: A pointer to a Vault struct containing the details of the newly created vault.
// - error: An error object if the operation fails.
func (cli *OpCLI) CreateVault(name, description string, icon VaultIcon, adminAccess bool) (*Vault, error) {
	// Validate the vault name
	if name == "" {
		return nil, errors.New("vault name cannot be empty")
	}

	// Execute the command to create a new vault
	output, err := cli.ExecuteOpCommand("vault", "create", name, "--description", description, "--icon", string(icon), "--allow-admins-to-manage", fmt.Sprintf("%t", adminAccess))
	if err != nil {
		return nil, err
	}

	var vault Vault
	err = json.Unmarshal(output, &vault)
	if err != nil {
		return nil, err
	}

	vault.cli = cli

	return &vault, nil
}

// ValidateVaultID validates the format of a vault ID.
//
// This method checks if the provided vault ID is a 26-character alphanumeric string.
//
// Parameters:
// - id: The vault ID to validate.
//
// Returns:
// - error: An error object if the ID format is invalid, otherwise nil.
func ValidateVaultID(id string) error {
	// Vault ID must be a 26-character alphanumeric string
	var validIDPattern = regexp.MustCompile(`^[a-z0-9]{26}$`)
	if !validIDPattern.MatchString(id) {
		return errors.New("invalid vault ID format")
	}
	return nil
}

// ValidateVault validates all fields of a Vault struct.
//
// This method performs comprehensive validation of a Vault struct, including checks for ID format, name, content version,
// timestamps, item count, description length, and type.
//
// Parameters:
// - vault: The Vault struct to validate.
//
// Returns:
// - error: An error object if any validation fails, otherwise nil.
func ValidateVault(vault Vault) error {
	if err := ValidateVaultID(vault.ID); err != nil {
		return err
	}

	if vault.Name == "" {
		return errors.New("vault name cannot be empty")
	}

	if vault.ContentVersion < 0 {
		return errors.New("content version must be a non-negative integer")
	}

	if _, err := time.Parse(time.RFC3339, vault.CreatedAt); err != nil {
		return errors.New("created_at must be a valid ISO 8601 date")
	}

	if _, err := time.Parse(time.RFC3339, vault.UpdatedAt); err != nil {
		return errors.New("updated_at must be a valid ISO 8601 date")
	}

	createdAt, _ := time.Parse(time.RFC3339, vault.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, vault.UpdatedAt)
	if updatedAt.Before(createdAt) {
		return errors.New("updated_at cannot be earlier than created_at")
	}

	if vault.Items < 0 {
		return errors.New("items must be a non-negative integer")
	}

	if len(vault.Description) > 500 {
		return errors.New("description cannot exceed 500 characters")
	}

	expectedTypes := map[string]bool{
		"USER_CREATED":     true,
		"SYSTEM_GENERATED": true,
	}
	if !expectedTypes[vault.Type] {
		return errors.New("type must be one of USER_CREATED or SYSTEM_GENERATED")
	}

	return nil
}

// UpdateVaultIcon updates the icon of a specified vault.
//
// This method validates the vault ID and icon name, then executes the "vault edit" command using the 1Password CLI
// to update the icon of the specified vault.
//
// Parameters:
// - vaultID: The unique identifier of the vault.
// - icon: The new icon to set for the vault. Must be a valid VaultIcon.
//
// Returns:
// - error: An error object if the operation fails.
func (cli *OpCLI) UpdateVaultIcon(vaultID string, icon VaultIcon) error {
	if err := ValidateVaultID(vaultID); err != nil {
		return err
	}

	validIcons := map[VaultIcon]bool{
		IconAirplane:         true,
		IconApplication:      true,
		IconArtSupplies:      true,
		IconBankersBox:       true,
		IconBrownBriefcase:   true,
		IconBrownGate:        true,
		IconBuildings:        true,
		IconCabin:            true,
		IconCastle:           true,
		IconCircleOfDots:     true,
		IconCoffee:           true,
		IconColorWheel:       true,
		IconCurtainedWindow:  true,
		IconDocument:         true,
		IconDoughnut:         true,
		IconFence:            true,
		IconGalaxy:           true,
		IconGears:            true,
		IconGlobe:            true,
		IconGreenBackpack:    true,
		IconGreenGem:         true,
		IconHandshake:        true,
		IconHeartWithMonitor: true,
		IconHouse:            true,
		IconIDCard:           true,
		IconJet:              true,
		IconLargeShip:        true,
		IconLuggage:          true,
		IconPlant:            true,
		IconPorthole:         true,
		IconPuzzle:           true,
		IconRainbow:          true,
		IconRecord:           true,
		IconRoundDoor:        true,
		IconSandals:          true,
		IconScales:           true,
		IconScrewdriver:      true,
		IconShop:             true,
		IconTallWindow:       true,
		IconTreasureChest:    true,
		IconVaultDoor:        true,
		IconVehicle:          true,
		IconWallet:           true,
		IconWrench:           true,
	}

	if !validIcons[icon] {
		return errors.New("invalid icon name")
	}

	_, err := cli.ExecuteOpCommand("vault", "edit", vaultID, "--icon", string(icon))
	if err != nil {
		return fmt.Errorf("failed to update vault icon: %w", err)
	}

	return nil
}

// GrantUserPermission grants a specific permission to a user for the current vault.
//
// This method validates the user and resolves the permission string, then executes the "vault user grant" command
// using the 1Password CLI to grant the specified permission to the user.
//
// Parameters:
// - user: The User struct representing the user to grant permission to.
// - permission: The Permission struct representing the permission to grant.
//
// Returns:
// - error: An error object if the operation fails.
func (vault *Vault) GrantUserPermission(user User, permission Permission) error {
	// Check if the user is valid
	if user.ID == "" {
		return errors.New("invalid user: user ID cannot be empty")
	}

	// Resolve the permission string using ResolvePermissions
	resolvedPermissions := ResolvePermissions(permission)

	// Execute the command to grant permissions
	_, err := vault.cli.ExecuteOpCommand(
		"vault", "user", "grant",
		"--vault", vault.ID,
		"--user", user.ID,
		"--permissions", resolvedPermissions,
	)
	if err != nil {
		return fmt.Errorf("failed to grant permissions: %w", err)
	}

	return nil
}

// RevokeUserPermission revokes a specific permission from a user for the current vault.
//
// This method validates the user and resolves the permission string, then executes the "vault user revoke" command
// using the 1Password CLI to revoke the specified permission from the user.
//
// Parameters:
// - user: The User struct representing the user to revoke permission from.
// - permission: The Permission struct representing the permission to revoke.
//
// Returns:
// - error: An error object if the operation fails.
func (vault *Vault) RevokeUserPermission(user User, permission Permission) error {
	// Check if the user is valid
	if user.ID == "" {
		return errors.New("invalid user: user ID cannot be empty")
	}

	// Resolve the permission string using ResolvePermissions
	resolvedPermissions := ResolvePermissions(permission)

	// Execute the command to revoke permissions
	_, err := vault.cli.ExecuteOpCommand(
		"vault", "user", "revoke",
		"--vault", vault.ID,
		"--user", user.ID,
		"--permissions", resolvedPermissions,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke permissions: %w", err)
	}

	return nil
}

// GrantGroupPermission grants a specific permission to a group for the current vault.
//
// This method validates the group and resolves the permission string, then executes the "vault group grant" command
// using the 1Password CLI to grant the specified permission to the group.
//
// Parameters:
// - group: The Group struct representing the group to grant permission to.
// - permission: The Permission struct representing the permission to grant.
//
// Returns:
// - error: An error object if the operation fails.
func (vault *Vault) GrantGroupPermission(group Group, permission Permission) error {
	// Check if the group is valid
	if group.ID == "" {
		return errors.New("invalid group: group ID cannot be empty")
	}

	// Resolve the permission string using ResolvePermissions
	resolvedPermissions := ResolvePermissions(permission)

	// Execute the command to grant permissions
	_, err := vault.cli.ExecuteOpCommand(
		"vault", "group", "grant",
		"--vault", vault.ID,
		"--group", group.ID,
		"--permissions", resolvedPermissions,
	)
	if err != nil {
		return fmt.Errorf("failed to grant permissions: %w", err)
	}

	return nil
}

// RevokeGroupPermission revokes a specific permission from a group for the current vault.
//
// This method validates the group and resolves the permission string, then executes the "vault group revoke" command
// using the 1Password CLI to revoke the specified permission from the group.
//
// Parameters:
// - group: The Group struct representing the group to revoke permission from.
// - permission: The Permission struct representing the permission to revoke.
//
// Returns:
// - error: An error object if the operation fails.
func (vault *Vault) RevokeGroupPermission(group Group, permission Permission) error {
	// Check if the group is valid
	if group.ID == "" {
		return errors.New("invalid group: group ID cannot be empty")
	}

	// Resolve the permission string using ResolvePermissions
	resolvedPermissions := ResolvePermissions(permission)

	// Execute the command to revoke permissions
	_, err := vault.cli.ExecuteOpCommand(
		"vault", "group", "revoke",
		"--vault", vault.ID,
		"--group", group.ID,
		"--permissions", resolvedPermissions,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke permissions: %w", err)
	}

	return nil
}

// Delete deletes the current vault.
//
// This method executes the "vault delete" command using the 1Password CLI to delete the current vault.
//
// Returns:
// - error: An error object if the operation fails.
func (vault *Vault) Delete() error {
	// Execute the command to delete the vault
	_, err := vault.cli.ExecuteOpCommand("vault", "delete", vault.ID)
	if err != nil {
		return fmt.Errorf("failed to delete vault: %w", err)
	}

	return nil
}

// SetName updates the name of the current vault.
//
// This method validates the new name and executes the "vault edit" command using the 1Password CLI to update the vault's name.
//
// Parameters:
// - name: The new name to set for the vault.
//
// Returns:
// - error: An error object if the operation fails.
func (vault *Vault) SetName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	args := []string{"vault", "edit", vault.ID, "--name", name}
	_, err := vault.cli.ExecuteOpCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to edit vault name: %w", err)
	}

	return nil
}

// SetDescription updates the description of the current vault.
//
// This method executes the "vault edit" command using the 1Password CLI to update the vault's description.
//
// Parameters:
// - description: The new description to set for the vault.
//
// Returns:
// - error: An error object if the operation fails.
func (vault *Vault) SetDescription(description string) error {
	args := []string{"vault", "edit", vault.ID, "--description", description}
	_, err := vault.cli.ExecuteOpCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to edit vault description: %w", err)
	}

	return nil
}

// SetIcon updates the icon of the current vault.
//
// This method validates the new icon and executes the "vault edit" command using the 1Password CLI to update the vault's icon.
//
// Parameters:
// - icon: The new icon to set for the vault. Must be a valid VaultIcon.
//
// Returns:
// - error: An error object if the operation fails.
func (vault *Vault) SetIcon(icon VaultIcon) error {
	if icon == "" {
		return errors.New("icon cannot be empty")
	}

	args := []string{"vault", "edit", vault.ID, "--icon", string(icon)}
	_, err := vault.cli.ExecuteOpCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to edit vault icon: %w", err)
	}

	return nil
}

// SetTravelMode sets the Travel Mode status for the current vault.
//
// This method executes the "vault edit" command using the 1Password CLI to update the Travel Mode status of the vault.
//
// Parameters:
// - travelModeOn: A boolean value indicating whether to turn Travel Mode on (true) or off (false).
//
// Returns:
// - error: An error object if the operation fails.
func (vault *Vault) SetTravelMode(travelModeOn bool) error {
	mode := "off"
	if travelModeOn {
		mode = "on"
	}

	args := []string{"vault", "edit", vault.ID, "--travel-mode", mode}
	_, err := vault.cli.ExecuteOpCommand(args...)
	if err != nil {
		return fmt.Errorf("failed to edit vault travel mode: %w", err)
	}

	return nil
}
