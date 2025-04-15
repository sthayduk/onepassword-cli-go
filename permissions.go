package onepassword

import "strings"

// Permission represents a specific permission in 1Password.
type Permission string

const (
	// Granular permissions
	PermissionViewItems            Permission = "view_items"
	PermissionCreateItems          Permission = "create_items"
	PermissionEditItems            Permission = "edit_items"
	PermissionArchiveItems         Permission = "archive_items"
	PermissionDeleteItems          Permission = "delete_items"
	PermissionViewAndCopyPasswords Permission = "view_and_copy_passwords"
	PermissionViewItemHistory      Permission = "view_item_history"
	PermissionImportItems          Permission = "import_items"
	PermissionExportItems          Permission = "export_items"
	PermissionCopyAndShareItems    Permission = "copy_and_share_items"
	PermissionPrintItems           Permission = "print_items"
	PermissionManageVault          Permission = "manage_vault"

	// Broader permissions
	PermissionAllowViewing  Permission = "allow_viewing"
	PermissionAllowEditing  Permission = "allow_editing"
	PermissionAllowManaging Permission = "allow_managing"

	// Derived permissions
	PermissionMoveItems Permission = "move_items"
)

// PermissionDependencies maps each permission to its required broader permissions.
type PermissionDependenciesMap map[Permission][]Permission

var PermissionDependencies = PermissionDependenciesMap{
	PermissionCreateItems:          {PermissionCreateItems, PermissionViewItems},
	PermissionViewAndCopyPasswords: {PermissionViewItems, PermissionViewAndCopyPasswords},
	PermissionEditItems:            {PermissionEditItems, PermissionViewAndCopyPasswords, PermissionViewItems},
	PermissionArchiveItems:         {PermissionArchiveItems, PermissionEditItems, PermissionViewAndCopyPasswords, PermissionViewItems},
	PermissionDeleteItems:          {PermissionDeleteItems, PermissionEditItems, PermissionViewAndCopyPasswords, PermissionViewItems},
	PermissionViewItemHistory:      {PermissionViewItemHistory, PermissionViewAndCopyPasswords, PermissionViewItems},
	PermissionImportItems:          {PermissionImportItems, PermissionCreateItems, PermissionViewItems},
	PermissionExportItems:          {PermissionExportItems, PermissionViewItemHistory, PermissionViewAndCopyPasswords, PermissionViewItems},
	PermissionCopyAndShareItems:    {PermissionCopyAndShareItems, PermissionViewItemHistory, PermissionViewAndCopyPasswords, PermissionViewItems},
	PermissionPrintItems:           {PermissionPrintItems, PermissionViewItemHistory, PermissionViewAndCopyPasswords, PermissionViewItems},
}

// ResolvePermissions generates a string of permissions for a given permission key in the PermissionDependenciesMap.
func ResolvePermissions(permission Permission) string {
	dependencies, exists := PermissionDependencies[permission]
	if !exists {
		return string(permission)
	}

	// Use a map to avoid duplicate permissions
	resolved := make(map[Permission]struct{})
	for _, dep := range dependencies {
		resolved[dep] = struct{}{}
	}

	// Convert the map keys to a comma-separated string
	var result []string
	for perm := range resolved {
		result = append(result, string(perm))
	}

	return strings.Join(result, ",")
}
