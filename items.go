package onepassword

import (
	"encoding/json"
	"fmt"
	"time"
)

// Category represents the type of item in 1Password
type Category string

const (
	CategoryLogin    Category = "LOGIN"
	CategoryPassword Category = "PASSWORD"
	CategorySecNote  Category = "SECURE_NOTE"
	CategoryIdentity Category = "IDENTITY"
)

// FieldType represents the type of a field
type FieldType string

const (
	FieldTypeString    FieldType = "STRING"
	FieldTypeConcealed FieldType = "CONCEALED"
)

// FieldPurpose represents the purpose of a field
type FieldPurpose string

const (
	PurposeUsername FieldPurpose = "USERNAME"
	PurposePassword FieldPurpose = "PASSWORD"
	PurposeNotes    FieldPurpose = "NOTES"
)

// PasswordStrength represents password strength levels
type PasswordStrength string

const (
	StrengthFantastic PasswordStrength = "FANTASTIC"
	StrengthTerrible  PasswordStrength = "TERRIBLE"
)

// ItemURL represents a URL associated with an item
type ItemURL struct {
	Href    string `json:"href"`
	Label   string `json:"label"`
	Primary bool   `json:"primary"`
}

// Section represents a section in an item
type Section struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// PasswordDetails contains password-specific information
type PasswordDetails struct {
	Strength PasswordStrength `json:"strength"`
	History  []string         `json:"history,omitempty"`
}

// Field represents a field in a 1Password item with its type, purpose, and value
type Field struct {
	ID              string           `json:"id"`
	Label           string           `json:"label"`
	Value           string           `json:"value,omitempty"`
	Reference       string           `json:"reference"`
	Type            FieldType        `json:"type"`
	Purpose         FieldPurpose     `json:"purpose,omitempty"`
	Section         *Section         `json:"section,omitempty"`
	PasswordDetails *PasswordDetails `json:"password_details,omitempty"`
}

// Item represents a 1Password item
type Item struct {
	cli *OpCLI `json:"-"` // Reference to the OpCLI instance for update operations

	ID             string    `json:"id"`
	Title          string    `json:"title"`
	LastEditedBy   string    `json:"last_edited_by"`
	AdditionalInfo string    `json:"additional_information"`
	Vault          Vault     `json:"vault"`
	Category       Category  `json:"category"`
	Favorite       bool      `json:"favorite"`
	Version        int       `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Tags           []string  `json:"tags,omitempty"`
	URLs           []ItemURL `json:"urls,omitempty"`
	Sections       []Section `json:"sections,omitempty"`
	Fields         []Field   `json:"fields,omitempty"`
}

// Check if a section ID is unique within the item
func (item *Item) isSectionIDUnique(sectionID string) bool {
	for _, sec := range item.Sections {
		if sec.ID == sectionID {
			return false
		}
	}
	return true
}

// AddSection adds a new section to the item.
//
// Parameters:
// - section: The Section struct to be added to the item.
//
// Returns:
// - error: An error object if the section ID is not unique.
//
// This method appends the provided section to the item's Sections slice.
func (item *Item) AddSection(section Section) error {
	if !item.isSectionIDUnique(section.ID) {
		return fmt.Errorf("SectionID is not unique within item")
	}
	item.Sections = append(item.Sections, section)
	return nil
}

// DeleteSection removes a section from the item by its ID.
//
// Parameters:
// - section: The Section struct to be removed from the item.
//
// This method ensures that all fields associated with the section are removed
// before deleting the section itself to maintain a consistent state.
func (item *Item) DeleteSection(section Section) {

	// Remove all fields associated with the section before deleting the section
	// This is important to avoid dangling references
	// and to ensure that the item is in a consistent state
	// when the section is deleted
	for _, field := range item.Fields {
		if field.Section != nil && field.Section.ID == section.ID {
			item.DeleteFieldFromSection(section, field)
		}
	}

	for i, sec := range item.Sections {
		if sec.ID == section.ID {
			item.Sections = append(item.Sections[:i], item.Sections[i+1:]...)
			break
		}
	}
}

// AddFieldToSection adds a field to a specific section in the item.
//
// Parameters:
// - section: The Section struct where the field will be added.
// - field: The Field struct to be added to the section.
//
// Returns:
// - error: An error object if the section is not found in the item.
//
// This method associates the field with the specified section and appends it
// to the item's Fields slice.
func (item *Item) AddFieldToSection(section Section, field Field) error {

	sectionFound := false

	for i, sec := range item.Sections {

		if sec.ID == section.ID && sec.Label == section.Label {
			sectionFound = true

			field.Section = &item.Sections[i]
			item.Fields = append(item.Fields, field)
			break
		}
	}

	if !sectionFound {
		return fmt.Errorf("Section not found in item")
	}

	return nil
}

// DeleteFieldFromSection removes a field from a specific section in the item.
//
// Parameters:
// - section: The Section struct from which the field will be removed.
// - field: The Field struct to be removed from the section.
//
// Returns:
// - error: An error object if the field is not found in the section.
//
// This method ensures that the field is properly disassociated from the section
// and removed from the item's Fields slice.
func (item *Item) DeleteFieldFromSection(section Section, field Field) error {

	itemFound := false

	for i, field := range item.Fields {
		if field.ID == field.ID && field.Section != nil && field.Section.ID == section.ID && field.Section.Label == section.Label {

			itemFound = true

			item.Fields = append(item.Fields[:i], item.Fields[i+1:]...)
			break
		}
	}

	if !itemFound {
		return fmt.Errorf("Field not found in section")
	}

	return nil
}

// ItemTemplate represents a 1Password item template
type ItemTemplate struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// GetItems retrieves a list of all items using the 1Password CLI.
//
// Returns:
// - *[]Item: A pointer to a slice of Item structs containing details of each item.
// - error: An error object if the operation fails.
//
// This method executes the "item list" command using the CLI and parses the
// JSON output into a slice of Item structs. It also populates the cli field
// for each item.
func (cli *OpCLI) GetItems() (*[]Item, error) {
	output, err := cli.ExecuteOpCommand("item", "list")
	if err != nil {
		return nil, err
	}

	var items []Item
	err = json.Unmarshal(output, &items)
	if err != nil {
		return nil, err
	}

	// Populate the cli field for each item
	for i := range items {
		items[i].cli = cli
	}

	return &items, nil
}

// getItem retrieves the details of a specific item by its identifier.
//
// Parameters:
// - identifier: A string representing the unique identifier of the item.
//
// Returns:
// - *Item: A pointer to the Item struct containing the item's details.
// - error: An error object if the operation fails.
//
// This method executes the "item get" command using the CLI and parses the
// JSON output into an Item struct. It also populates the cli field for the item.
func (cli *OpCLI) getItem(identifier string) (*Item, error) {

	output, err := cli.ExecuteOpCommand("item", "get", identifier)
	if err != nil {
		return nil, err
	}

	var item Item
	err = json.Unmarshal(output, &item)
	if err != nil {
		return nil, err
	}

	// Populate the cli field for the item
	item.cli = cli

	return &item, nil
}

// GetItemByName retrieves an item by its name.
//
// Parameters:
// - itemName: A string representing the name of the item.
//
// Returns:
// - *Item: A pointer to the Item struct containing the item's details.
// - error: An error object if the operation fails.
func (cli *OpCLI) GetItemByName(itemName string) (*Item, error) {
	return cli.getItem(itemName)
}

// GetItemByID retrieves an item by its ID.
//
// Parameters:
// - itemID: A string representing the unique identifier of the item.
//
// Returns:
// - *Item: A pointer to the Item struct containing the item's details.
// - error: An error object if the operation fails.
func (cli *OpCLI) GetItemByID(itemID string) (*Item, error) {
	return cli.getItem(itemID)
}

// GetItemTemplateByName retrieves an item template by its name.
//
// Parameters:
// - templateName: An ItemTemplate struct representing the template to retrieve.
//
// Returns:
// - *Item: A pointer to the Item struct containing the template's details.
// - error: An error object if the operation fails.
func (cli *OpCLI) GetItemTemplateByName(templateName ItemTemplate) (*Item, error) {
	return cli.getItem(templateName.Name)
}

// GetItemTemplates retrieves a list of all item templates using the 1Password CLI.
//
// Returns:
// - *[]ItemTemplate: A pointer to a slice of ItemTemplate structs containing details of each template.
// - error: An error object if the operation fails.
//
// This method executes the "item template list" command using the CLI and parses
// the JSON output into a slice of ItemTemplate structs.
func (cli *OpCLI) GetItemTemplates() (*[]ItemTemplate, error) {
	output, err := cli.ExecuteOpCommand("item", "template", "list")
	if err != nil {
		return nil, err
	}

	var itemTemplates []ItemTemplate
	err = json.Unmarshal(output, &itemTemplates)
	if err != nil {
		return nil, err
	}

	return &itemTemplates, nil
}
