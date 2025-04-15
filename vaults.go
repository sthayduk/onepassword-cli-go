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
	cli *OpCLI `json:"-"` // Reference to the OpCLI instance for update operations	// Exclude cli from JSON serialization

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
// Returns:
// - []Vault: A slice of Vault structs containing details of each vault.
// - error: An error object if the operation fails.
func (cli *OpCLI) GetVaultDetails() ([]Vault, error) {
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

	return vaults, nil
}

// getVaultDetails retrieves the details of a specific vault by its identifier.
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
	// This is necessary for operations that require the cli context
	// such as updating the vault icon
	// or any other operations that may be added in the future
	vault.cli = cli

	return &vault, nil
}

// GetVaultDetailsByName retrieves the details of a vault by its name.
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

// ValidateVaultID validates the format of a vault ID.
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

// CreateVault creates a new vault in the 1Password account using the provided parameters.
//
// Parameters:
//   - name: The name of the vault to be created. This parameter is required and cannot be empty.
//   - description: A brief description of the vault. The description must not exceed 500 characters.
//   - icon: The icon to represent the vault. This should be a value of type VaultIcon.
//   - adminAccess: A boolean flag indicating whether administrators are allowed to manage the vault.
//     If set to true, administrators will have management access to the vault.
//     If set to false, administrators will not have management access.
//
// Returns:
//   - *Vault: A pointer to the created Vault object containing details of the newly created vault.
//   - error: An error object if the vault creation fails or if there are issues with the input parameters.
//
// Errors:
//   - Returns an error if the vault name is empty.
//   - Returns an error if the description exceeds 500 characters.
//   - Returns an error if the underlying `op` CLI command fails to execute.
//   - Returns an error if the output from the `op` CLI command cannot be parsed into a Vault object.
//
// Example:
//
//	vault, err := cli.CreateVault("MyVault", "This is a secure vault", VaultIconKey, true)
//	if err != nil {
//	    log.Fatalf("Error creating vault: %v", err)
//	}
//	fmt.Printf("Vault created successfully: %+v\n", vault)
func (cli *OpCLI) CreateVault(name, description string, icon VaultIcon, adminAccess bool) (*Vault, error) {
	if name == "" {
		return nil, errors.New("vault name cannot be empty")
	}

	if len(description) > 500 {
		return nil, errors.New("description cannot exceed 500 characters")
	}

	output, err := cli.ExecuteOpCommand("vault", "create", "--name", name, "--description", description, "--allow-admins-to-manage", fmt.Sprintf("%t", adminAccess), "--icon", string(icon))
	if err != nil {
		return nil, fmt.Errorf("failed to create vault: %w", err)
	}

	var vault Vault
	err = json.Unmarshal(output, &vault)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vault details: %w", err)
	}

	vault.cli = cli

	return &vault, nil
}

// DeleteVault deletes a specified vault from the 1Password account.
//
// Parameters:
// - vault: The Vault object representing the vault to be deleted.
//
// Returns:
// - error: An error object if the operation fails, otherwise nil.
//
// Example:
//
//	err := cli.DeleteVault(vault)
//	if err != nil {
//	    log.Fatalf("Failed to delete vault: %v", err)
//	}
func (cli *OpCLI) DeleteVault(vault Vault) error {
	if err := ValidateVaultID(vault.ID); err != nil {
		return err
	}

	_, err := cli.ExecuteOpCommand("vault", "delete", vault.ID)
	if err != nil {
		return fmt.Errorf("failed to delete vault: %w", err)
	}

	return nil
}

// SetIcon updates the icon of the vault.
//
// Parameters:
// - icon: The new icon to set for the vault. Must be a valid VaultIcon.
//
// Returns:
// - error: An error object if the operation fails, otherwise nil.
//
// Example:
//
//	err := vault.SetIcon(IconHeartWithMonitor)
//	if err != nil {
//	    log.Fatalf("Failed to update vault icon: %v", err)
//	}
func (vault *Vault) SetIcon(icon VaultIcon) error {

	if err := ValidateVaultID(vault.ID); err != nil {
		return err
	}

	err := vault.cli.UpdateVaultIcon(vault.ID, icon)
	if err != nil {
		return fmt.Errorf("failed to update vault icon: %w", err)
	}

	return nil
}

// Delete removes the vault from the 1Password account.
//
// Returns:
// - error: An error object if the operation fails, otherwise nil.
//
// Example:
//
//	err := vault.Delete()
//	if err != nil {
//	    log.Fatalf("Failed to delete vault: %v", err)
//	}
func (vault *Vault) Delete() error {
	if err := ValidateVaultID(vault.ID); err != nil {
		return err
	}

	err := vault.cli.DeleteVault(*vault)
	if err != nil {
		return fmt.Errorf("failed to delete vault: %w", err)
	}

	return nil
}

// SetName updates the name of the vault.
//
// Parameters:
// - name: The new name for the vault. Must not be empty.
//
// Returns:
// - error: An error object if the operation fails, otherwise nil.
//
// Example:
//
//	err := vault.SetName("NewVaultName")
//	if err != nil {
//	    log.Fatalf("Failed to update vault name: %v", err)
//	}
func (vault *Vault) SetName(name string) error {
	if name == "" {
		return errors.New("vault name cannot be empty")
	}

	vault.Name = name

	_, err := vault.cli.ExecuteOpCommand("vault", "edit", vault.ID, "--name", name)
	if err != nil {
		return fmt.Errorf("failed to update vault name: %w", err)
	}

	return nil
}

// SetDescription updates the description of the vault.
//
// Parameters:
// - description: The new description for the vault. Must not be empty and must not exceed 500 characters.
//
// Returns:
// - error: An error object if the operation fails, otherwise nil.
//
// Example:
//
//	err := vault.SetDescription("This is a new description for the vault.")
//	if err != nil {
//	    log.Fatalf("Failed to update vault description: %v", err)
//	}
func (vault *Vault) SetDescription(description string) error {
	if description == "" {
		return errors.New("vault description cannot be empty")
	}
	if len(description) > 500 {
		return errors.New("description cannot exceed 500 characters")
	}
	vault.Description = description
	_, err := vault.cli.ExecuteOpCommand("vault", "edit", vault.ID, "--description", description)
	if err != nil {
		return fmt.Errorf("failed to update vault description: %w", err)
	}

	return nil
}

// SetTravelMode updates the Travel Mode setting for the vault.
//
// Travel Mode is a feature that allows users to specify which vaults are accessible
// when Travel Mode is turned on. Vaults with Travel Mode enabled will remain accessible,
// while others will be hidden. This is particularly useful for securely traveling across
// borders or through areas where sensitive data might be at risk.
//
// The method takes a boolean parameter `travelMode`:
// - If `true`, Travel Mode is enabled for the vault.
// - If `false`, Travel Mode is disabled for the vault.
//
// The method internally executes the `op` CLI command to update the vault's Travel Mode setting:
// `op vault edit <vault-id> --travel-mode <on|off>`
//
// Parameters:
// - travelMode (bool): A flag indicating whether to enable or disable Travel Mode.
//
// Returns:
// - error: An error if the operation fails, or `nil` if the operation succeeds.
//
// Example:
//
//	err := vault.SetTravelMode(true)
//	if err != nil {
//	    log.Fatalf("Failed to enable Travel Mode: %v", err)
//	}
func (vault *Vault) SetTravelMode(travelMode bool) error {
	_, err := vault.cli.ExecuteOpCommand("vault", "edit", vault.ID, "--travel-mode", fmt.Sprintf("%t", travelMode))
	if err != nil {
		return fmt.Errorf("failed to update vault travel mode: %w", err)
	}

	return nil
}
