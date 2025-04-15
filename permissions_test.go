package onepassword

import (
	"strings"
	"testing"
)

func TestResolvePermissions(t *testing.T) {
	tests := []struct {
		name       string
		permission Permission
		expected   string
	}{
		{
			name:       "Permission with no dependencies",
			permission: PermissionManageVault,
			expected:   "manage_vault",
		},
		{
			name:       "Permission with single dependency",
			permission: PermissionCreateItems,
			expected:   "create_items,view_items",
		},
		{
			name:       "Permission with multiple dependencies",
			permission: PermissionEditItems,
			expected:   "edit_items,view_and_copy_passwords,view_items",
		},
		{
			name:       "Permission not in dependencies map",
			permission: PermissionMoveItems,
			expected:   "view_items,edit_items,archive_items,view_and_copy_passwords,view_item_history,copy_and_share_items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolvePermissions(tt.permission)
			// Split and sort to ensure order doesn't affect comparison
			resultParts := strings.Split(result, ",")
			expectedParts := strings.Split(tt.expected, ",")
			sortStrings(resultParts)
			sortStrings(expectedParts)

			if strings.Join(resultParts, ",") != strings.Join(expectedParts, ",") {
				t.Errorf("ResolvePermissions(%q) = %q; want %q", tt.permission, result, tt.expected)
			}
		})
	}
}

// Helper function to sort a slice of strings
func sortStrings(slice []string) {
	for i := 0; i < len(slice); i++ {
		for j := i + 1; j < len(slice); j++ {
			if slice[i] > slice[j] {
				slice[i], slice[j] = slice[j], slice[i]
			}
		}
	}
}
