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
	// This is necessary for operations that require the cli context
	// such as updating the vault icon
	// or any other operations that may be added in the future
	for i := range vaults {
		vaults[i].cli = cli
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
	output, err := cli.ExecuteOpCommand("vault", "create", "--name", name, "--description", description, "--icon", string(icon), "--allow-admins-to-manage", fmt.Sprintf("%t", adminAccess))
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
