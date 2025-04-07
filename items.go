package onepassword

import "time"

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
	ID string `json:"id"`
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
