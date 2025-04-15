package onepassword

import (
	"encoding/json"
	"fmt"
	"slices"
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
	FieldTypeString    FieldType = "STRING"     // A text string.
	FieldTypeConcealed FieldType = "CONCEALED"  // A concealed password.
	FieldTypeEmail     FieldType = "EMAIL"      // An email address.
	FieldTypeURL       FieldType = "URL"        // A web address to copy or open in your default web browser, not used for autofill behavior. Use the --url flag to set the website where 1Password suggests and fills a Login, Password, or API Credential item.
	FieldTypeDate      FieldType = "DATE"       // A date with the format YYYY-MM-DD.
	FieldTypeMonthYear FieldType = "MONTH_YEAR" // A date with the format YYYYMM or YYYY/MM.
	FieldTypePhone     FieldType = "PHONE"      // A phone number.
	FieldTypeOTP       FieldType = "OTP"        // A one-time password. Accepts an otpauth:// URI as the value.
	FieldTypeFile      FieldType = "N/A"        // A file attachment. Accepts the path to the file as the value. Can only be added with assignment statements.
)

// FieldPurpose represents the purpose of a field
type FieldPurpose string

const (
	FieldPurposeUsername  FieldPurpose = "username"  // A username.
	FieldPurposeNotes     FieldPurpose = "notes"     // A notes field.
	FieldPurposePassword  FieldPurpose = "password"  // A concealed password.
	FieldPurposeText      FieldPurpose = "text"      // A text string.
	FieldPurposeEmail     FieldPurpose = "email"     // An email address.
	FieldPurposeURL       FieldPurpose = "url"       // A web address to copy or open in your default web browser, not used for autofill behavior. Use the --url flag to set the website where 1Password suggests and fills a Login, Password, or API Credential item.
	FieldPurposeDate      FieldPurpose = "date"      // A date with the format YYYY-MM-DD.
	FieldPurposeMonthYear FieldPurpose = "monthyear" // A date with the format YYYYMM or YYYY/MM.
	FieldPurposePhone     FieldPurpose = "phone"     // A phone number.
	FieldPurposeOTP       FieldPurpose = "otp"       // A one-time password. Accepts an otpauth:// URI as the value.
	FieldPurposeFile      FieldPurpose = "file"      // A file attachment. Accepts the path to the file as the value. Can only be added with assignment statements.
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
	Strength  PasswordStrength `json:"strength"`
	History   []string         `json:"history,omitempty"`
	Entropy   float64          `json:"entropy,omitempty"`
	Generated bool             `json:"generated,omitempty"`
}

// Field represents a field in a 1Password item with its type, purpose, and value
type Field struct {
	ID              string           `json:"id,omitempty"`
	Label           string           `json:"label"`
	Value           string           `json:"value,omitempty"`
	Reference       string           `json:"reference,omitempty"`
	Type            FieldType        `json:"type"`
	Purpose         FieldPurpose     `json:"purpose,omitempty"`
	Section         *Section         `json:"section,omitempty"`
	PasswordDetails *PasswordDetails `json:"password_details,omitempty"`
	Entropy         float64          `json:"entropy,omitempty"`
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

// ToJSON converts the Item struct into a JSON-encoded byte slice.
// It returns the JSON representation of the item or an error if the
// marshaling process fails.
func (item *Item) ToJSON() ([]byte, error) {
	return json.Marshal(item)
}

// AddUserName adds or updates a username field in the item.
//
// Parameters:
// - username: A string representing the username to add or update.
//
// This method checks if a username field already exists in the item. If it does,
// it updates the value of the existing field. Otherwise, it creates a new username field
// and appends it to the item's Fields slice.
func (item *Item) AddUserName(username string) {
	// Check if a username field already exists and update it
	for i, field := range item.Fields {
		if field.Purpose == FieldPurposeUsername && field.Section != nil {
			item.Fields[i].Value = username
			return
		}
	}

	// If no username field exists, create and add a new one
	newField := Field{
		ID:      "username",
		Type:    FieldTypeString,
		Purpose: FieldPurposeUsername,
		Label:   "username",
		Value:   username,
	}

	item.Fields = append(item.Fields, newField)
}

// AddPassword adds or updates a password field in the item.
//
// Parameters:
// - password: A string representing the password to add or update.
//
// This method checks if a password field already exists in the item. If it does,
// it updates the value of the existing field. Otherwise, it creates a new password field
// and appends it to the item's Fields slice.
func (item *Item) AddPassword(password string) {
	// Check if a password field already exists and update it
	for i, field := range item.Fields {
		if field.Purpose == FieldPurposePassword && field.Section != nil {
			item.Fields[i].Value = password
			return
		}
	}
	// If no password field exists, create and add a new one
	newField := Field{
		ID:      "password",
		Type:    FieldTypeConcealed,
		Purpose: FieldPurposePassword,
		Label:   "password",
		Value:   password,
	}
	item.Fields = append(item.Fields, newField)
}

// AddNotes adds or updates a notes field in the item.
//
// Parameters:
// - notes: A string representing the notes to add or update.
//
// This method checks if a notes field already exists in the item. If it does,
// it updates the value of the existing field. Otherwise, it creates a new notes field
// and appends it to the item's Fields slice.
func (item *Item) AddNotes(notes string) {
	// Check if a notes field already exists and update it
	for i, field := range item.Fields {
		if field.Purpose == FieldPurposeNotes && field.Section != nil {
			item.Fields[i].Value = notes
			return
		}
	}
	// If no notes field exists, create and add a new one
	newField := Field{
		ID:      "notes",
		Type:    FieldTypeString,
		Purpose: FieldPurposeNotes,
		Label:   "notes",
		Value:   notes,
	}
	item.Fields = append(item.Fields, newField)
}

// GetFieldByID retrieves a field by its ID.
//
// Parameters:
// - fieldID: A string representing the unique identifier of the field.
//
// Returns:
// - *Field: A pointer to the Field struct if found.
// - error: An error object if the field is not found.
func (item *Item) GetFieldByID(fieldID string) (*Field, error) {
	for _, field := range item.Fields {
		if field.ID == fieldID {
			return &field, nil
		}
	}
	return nil, fmt.Errorf("Field with ID '%s' not found", fieldID)
}

// SetFavorite sets the favorite status of the item.
// It updates the Favorite field of the Item struct to the specified boolean value.
//
// Parameters:
//   - favorite: A boolean value indicating whether the item should be marked as a favorite.
func (item *Item) SetAsFavorite(favorite bool) {
	item.Favorite = favorite
}

// GetFieldsByLabel retrieves fields by their label.
//
// Parameters:
// - fieldLabel: A string representing the label of the fields to retrieve.
//
// Returns:
// - []*Field: A slice of pointers to Field structs matching the label.
// - error: An error object if no fields with the given label are found.
func (item *Item) GetFieldsByLabel(fieldLabel string) ([]*Field, error) {
	var fields []*Field

	for _, field := range item.Fields {
		if field.Label == fieldLabel {
			fields = append(fields, &field)
		}
	}

	if len(fields) > 0 {
		return fields, nil
	}

	return nil, fmt.Errorf("Field with Label '%s' not found", fieldLabel)
}

// GetFieldsByPurpose retrieves fields by their purpose.
//
// Parameters:
// - fieldPurpose: A FieldPurpose value representing the purpose of the fields to retrieve.
//
// Returns:
// - []*Field: A slice of pointers to Field structs matching the purpose.
// - error: An error object if no fields with the given purpose are found.
func (item *Item) GetFieldsByPurpose(fieldPurpose FieldPurpose) ([]*Field, error) {
	var fields []*Field
	for _, field := range item.Fields {
		if field.Purpose == fieldPurpose {
			fields = append(fields, &field)
		}
	}

	if len(fields) > 0 {
		return fields, nil
	}

	return nil, fmt.Errorf("Field with Purpose '%s' not found", fieldPurpose)
}

// NewField creates a new Field instance with the specified label, value, and type.
//
// Parameters:
// - label: A string representing the label of the field.
// - value: A string representing the value of the field.
// - fieldType: A FieldType value representing the type of the field.
//
// Returns:
// - Field: A new Field struct initialized with the provided parameters.
func (item *Item) NewField(label, value string, fieldType FieldType) Field {
	return Field{
		Label: label,
		Value: value,
		Type:  fieldType,
	}
}

// AddField appends a new field to the item's Fields slice.
//
// Parameters:
// - field: The Field struct to be added to the item.
func (item *Item) AddField(field Field) {
	item.Fields = append(item.Fields, field)
}

// DeleteField removes a field from the item by its ID.
//
// Parameters:
// - field: The Field struct to be removed from the item.
//
// Returns:
// - error: An error object if the field with the specified ID is not found.
func (item *Item) DeleteField(field Field) error {
	if len(item.Fields) == 0 {
		return fmt.Errorf("No fields found to remove")
	}

	for i, f := range item.Fields {
		if f.ID == field.ID {
			item.Fields = append(item.Fields[:i], item.Fields[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Field with ID '%s' not found", field.ID)
}

// DeleteTag removes a tag from the item by its name.
//
// Parameters:
// - tag: A string representing the name of the tag to remove.
//
// Returns:
// - error: An error object if the tag with the specified name is not found.
func (item *Item) DeleteTag(tag string) error {
	if len(item.Tags) == 0 {
		return fmt.Errorf("No tags found to remove")
	}

	for i, t := range item.Tags {
		if t == tag {
			item.Tags = append(item.Tags[:i], item.Tags[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Tag '%s' not found", tag)
}

// AddTag appends a new tag to the item's Tags slice.
//
// Parameters:
// - tag: A string representing the tag to add.
func (item *Item) AddTag(tag string) {
	item.Tags = append(item.Tags, tag)
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
func (item *Item) DeleteSection(section Section) error {

	// Remove all fields associated with the section before deleting the section
	// This is important to avoid dangling references
	// and to ensure that the item is in a consistent state
	// when the section is deleted
	for _, field := range item.Fields {
		if field.Section != nil && field.Section.ID == section.ID {

			err := item.DeleteFieldFromSection(section, field)
			if err != nil {
				return err
			}
		}
	}

	for i, sec := range item.Sections {
		if sec.ID == section.ID {
			item.Sections = slices.Delete(item.Sections, i, i+1)
			break
		}
	}

	return nil
}

// AddFieldToSection adds a new field to a specific section in the item.
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

// MoveFieldToSection moves a field to a specific section in the item.
//
// Parameters:
// - field: The Field struct to be moved.
// - section: The Section struct where the field will be moved.
//
// Returns:
// - error: An error object if the section is not found in the item.
//
// This method associates the field with the specified section and updates its Section reference.
// If the section is not found, it returns an error.
func (item *Item) MoveFieldToSection(field Field, section Section) error {
	// Find the section in the item
	var foundSection *Section
	for i := range item.Sections {
		if item.Sections[i].ID == section.ID && item.Sections[i].Label == section.Label {
			foundSection = &item.Sections[i]
			break
		}
	}

	if foundSection == nil {
		return fmt.Errorf("Section not found in item")
	}

	// Find the field in the item and update its section
	for i := range item.Fields {
		if item.Fields[i].ID == field.ID {
			item.Fields[i].Section = foundSection
			return nil
		}
	}

	return fmt.Errorf("Field not found in item")
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

// Save saves the current state of the item to the 1Password CLI.
//
// Returns:
// - error: An error object if the operation fails.
//
// This method uses the UpdateItemWithStruct method of the OpCLI instance to
// save the item. It ensures that the cli field and item ID are properly set
// before attempting to save.
func (item *Item) Save() error {
	if item.cli == nil {
		return fmt.Errorf("cli is nil, cannot save item")
	}
	if item.ID == "" {
		return fmt.Errorf("item ID is empty, cannot save item")
	}

	// Use the new UpdateItemWithStruct method to save the item
	item, err := item.cli.updateItemWithStruct(*item)
	if err != nil {
		return fmt.Errorf("failed to save item: %v", err)
	}

	return nil
}

// Delete deletes the item from the 1Password CLI.
//
// Returns:
// - error: An error object if the operation fails.
//
// This method uses the DeleteItem method of the OpCLI instance to delete the
// item. It ensures that the cli field and item ID are properly set before
// attempting to delete.
func (item *Item) Delete() error {
	if item.cli == nil {
		return fmt.Errorf("cli is nil, cannot delete item")
	}
	if item.ID == "" {
		return fmt.Errorf("item ID is empty, cannot delete item")
	}

	// Use the new DeleteItem method to delete the item
	if err := item.cli.deleteItem(*item); err != nil {
		return fmt.Errorf("failed to delete item: %v", err)
	}
	return nil
}

// AddURL adds a new ItemURL to the item.
//
// Parameters:
// - url: The ItemURL struct to be added to the item.
//
// This method appends the provided URL to the item's URLs slice. If the URL
// is marked as primary, it ensures no other URL is marked as primary.
func (item *Item) AddURL(url ItemURL) {
	if url.Primary {
		// Ensure no other URL is marked as primary
		for i := range item.URLs {
			item.URLs[i].Primary = false
		}
	}
	item.URLs = append(item.URLs, url)
}

// DeleteURLs removes all ItemURLs from the item that match the given Href.
//
// Parameters:
// - href: A string representing the Href of the URLs to remove.
//
// Returns:
// - error: An error object if no URLs with the given Href are found or if the last URL cannot be deleted.
//
// Note: The 1Password CLI has a known issue where the last URL cannot be deleted. This method will
// return an error if attempting to delete the last remaining URL.
func (item *Item) DeleteURLs(href string) error {
	if len(item.URLs) == 0 {
		return fmt.Errorf("no URLs found to remove")
	}

	if len(item.URLs) == 1 {
		return fmt.Errorf("cannot delete the last URL due to a known issue in the 1Password CLI")
	}

	updatedURLs := item.URLs[:0] // Create a new slice to hold non-matching URLs
	found := false

	for _, url := range item.URLs {
		if url.Href == href {
			found = true
			continue // Skip URLs that match the Href
		}
		updatedURLs = append(updatedURLs, url)
	}

	if !found {
		return fmt.Errorf("no URLs with href '%s' found", href)
	}

	item.URLs = updatedURLs
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
// - templateName: A string representing the name of the template.
//
// Returns:
// - *Item: A pointer to the Item struct containing the template's details.
// - error: An error object if the operation fails.
//
// This method executes the "item template get" command using the CLI and parses
// the JSON output into an Item struct. It also populates the cli field for the item.
func (cli *OpCLI) GetItemTemplateByName(templateName string) (*Item, error) {
	output, err := cli.ExecuteOpCommand("item", "template", "get", templateName)
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
