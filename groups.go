package onepassword

import "time"

type Group struct {
	cli *OpCLI `json:"-"` // Reference to the OpCLI instance for update operations

	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Permissions []string  `json:"permissions,omitempty"`
	Type        string    `json:"type"`
}
