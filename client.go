package onepassword

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// OpCLI represents the 1Password CLI executor
type OpCLI struct {
	Path        string
	accesstoken string
	cache       itemCache
	logger      slog.Logger
	Account     *Account
}

// OpCliError represents an error from the 1Password CLI operations
type OpCliError struct {
	StderrOutput string
	Err          error
}

// Error returns the string representation of the CLI error
func (e *OpCliError) Error() string {
	if e.StderrOutput != "" {
		return e.StderrOutput
	}
	return e.Err.Error()
}

// itemCache maintains a local cache of 1Password items for faster lookups
type itemCache struct {
	items       map[string]*Item // key is item title
	initialized bool
}

// NewOpCLI initializes a new instance of the OpCLI struct.
// It locates the 1Password CLI executable and sets up an empty item cache.
//
// Returns:
// - A pointer to an OpCLI instance.
func NewOpCLI() *OpCLI {

	// Find the 1Password CLI executable
	opPath, err := FindOpExecutable()
	if err != nil {
		slog.Error("1Password CLI not found", "error", err)
	}

	return &OpCLI{
		Path:  opPath,
		cache: itemCache{items: make(map[string]*Item)},
	}
}

// FindOpExecutable searches for the "op" executable in the system's PATH.
// It iterates through each directory in the PATH environment variable and checks
// if the "op" executable exists and is not a directory. On Windows, it appends
// ".exe" to the executable name.
//
// Returns:
// - The full path to the "op" executable if found.
// - An error if the executable is not found in any of the directories in PATH.
func FindOpExecutable() (string, error) {
	slog.Debug("searching for op executable in PATH")
	paths := filepath.SplitList(os.Getenv("PATH"))
	executableName := "op"
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}
	for _, path := range paths {
		opPath := filepath.Join(path, executableName)
		if stat, err := os.Stat(opPath); err == nil && !stat.IsDir() {
			return opPath, nil
		}
	}
	return "", fmt.Errorf("no valid op executable found in PATH")
}

// TestOpCli verifies the availability of the 1Password CLI by executing the
// "--version" command using the provided path to the CLI executable.
// It returns an error if the command fails to execute or the CLI is not found.
//
// Parameters:
//   - opPath: The file path to the 1Password CLI executable.
//
// Returns:
//   - error: An error if the CLI is unavailable or the command execution fails,
//     otherwise nil.
func TestOpCli(opPath string) error {
	if err := exec.Command(opPath, "--version").Run(); err != nil {
		return err
	}
	return nil
}

// VerifyOpExecutable verifies the digital signature of the specified executable file
// to ensure it is signed by AgileBits Inc. (the developers of 1Password).
//
// The verification process varies depending on the operating system:
//   - On macOS, it uses the `codesign` command to check the signature authority.
//   - On Windows, it uses PowerShell's `Get-AuthenticodeSignature` to validate the signature
//     and ensure it is signed by AgileBits.
//   - On Linux, it uses GPG to verify the signature against AgileBits' public GPG key.
//
// If the verification fails or the executable is not signed by AgileBits, an error is returned.
// On Linux, if GPG is not available, the function logs a warning and skips the verification.
//
// Parameters:
//   - path: The file path to the executable to be verified.
//
// Returns:
//   - An error if the verification fails or the executable is not signed by AgileBits.
//     Returns nil if the verification succeeds or is skipped (e.g., on Linux without GPG).
func VerifyOpExecutable(path string) error {
	if runtime.GOOS == "darwin" {
		// Use codesign on macOS
		cmd := exec.Command("codesign", "-d", "-vvv", path)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("signature verification failed: %v", err)
		}

		// Check for AgileBits signature
		outputStr := string(output)
		if !strings.Contains(outputStr, "Authority=Developer ID Application: AgileBits Inc.") {
			return fmt.Errorf("invalid signature: not signed by AgileBits")
		}
	} else if runtime.GOOS == "windows" {
		// Use PowerShell Get-AuthenticodeSignature on Windows with publisher check
		cmd := exec.Command("powershell", "-Command", fmt.Sprintf(`
			$sig = Get-AuthenticodeSignature '%s'
			if ($sig.Status -ne 'Valid') { 
				exit 1 
			}
			if ($sig.SignerCertificate.Subject -notlike '*AgileBits*') {
				exit 1
			}`, path))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("signature verification failed or not signed by AgileBits: %v", err)
		}
	} else if runtime.GOOS == "linux" {
		// Check if gpg is available
		if _, err := exec.LookPath("gpg"); err != nil {
			slog.Warn("gpg not found, skipping signature verification on Linux")
			return nil
		}

		// AgileBits GPG key ID for 1Password packages
		const agileBitsKeyID = "3FEF9748469ADBE15DA7CA80AC2D62742012EA22"

		// Verify if key is in keyring, if not fetch it
		checkKey := exec.Command("gpg", "--list-keys", agileBitsKeyID)
		if err := checkKey.Run(); err != nil {
			slog.Info("fetching AgileBits GPG key")
			fetchKey := exec.Command("gpg", "--keyserver", "keyserver.ubuntu.com", "--recv-keys", agileBitsKeyID)
			if err := fetchKey.Run(); err != nil {
				return fmt.Errorf("failed to fetch AgileBits GPG key: %v", err)
			}
		}

		// Verify signature
		cmd := exec.Command("gpg", "--verify", path)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("GPG signature verification failed: %v", err)
		}

		// Verify it's signed by AgileBits
		verify := exec.Command("gpg", "--verify", "--debug-level", "1", path)
		output, _ := verify.CombinedOutput()
		if !strings.Contains(string(output), agileBitsKeyID) {
			return fmt.Errorf("invalid signature: not signed by AgileBits")
		}
	}
	return nil
}

// SignIn attempts to sign in to a 1Password account using the provided account details.
// It first tries a passwordless sign-in method. If that fails and the error indicates
// that password authentication is required, it prompts the user for a password and
// retries the sign-in process.
//
// Upon successful sign-in, the session token is stored in an environment variable
// and the account's sign-in information is updated.
//
// Parameters:
//   - ctx: The context for managing the command execution lifecycle.
//   - account: A pointer to the Account struct containing the account details.
//
// Returns:
//   - An error if the sign-in process fails, or nil if the sign-in is successful.
func (cli *OpCLI) SignIn(ctx context.Context, account *Account) error {
	slog.Debug("attempting to sign in to 1Password")

	slog.Debug("signing in to account",
		"account", account.UserUUID,
		"email", account.Email)

	signinCmd := exec.CommandContext(ctx, cli.Path, "signin", "--account", account.UserUUID, "--raw")
	var stderr bytes.Buffer
	var stdout bytes.Buffer
	signinCmd.Stderr = &stderr
	signinCmd.Stdout = &stdout

	var sessionToken string
	slog.Debug("attempting passwordless signin")
	err := signinCmd.Run()
	if err == nil {
		sessionToken = strings.TrimSpace(stdout.String())
		if sessionToken != "" {
			if err := os.Setenv("OP_SESSION_"+account.UserUUID, sessionToken); err != nil {
				return fmt.Errorf("failed to set session token: %v", err)
			}
		}

		slog.Debug("passwordless signin successful", "sessionToken", sessionToken)
		account.SetSignInInfo(sessionToken)
		cli.Account = account

		slog.Info("connected to 1Password", "url", account.URL, "email", account.Email)
		return nil
	}

	stderrOutput := stderr.String()
	slog.Debug("initial signin attempt failed", "error", err, "stderr", stderrOutput)

	if strings.Contains(strings.ToLower(stderrOutput), "enter the password for") ||
		strings.Contains(strings.ToLower(stderrOutput), "authentication") {

		slog.Debug("password authentication required")
		password, err := readPassword()
		if err != nil {
			return fmt.Errorf("error reading password: %v", err)
		}

		cmd := exec.CommandContext(ctx, cli.Path, "signin", "--account", account.UserUUID, "--raw")
		cmd.Stdin = strings.NewReader(password)
		output, err := cmd.Output()
		if err != nil {
			slog.Error("password signin failed", "error", err)
			return fmt.Errorf("signin failed: %v", err)
		}

		sessionToken = strings.TrimSpace(string(output))
		if sessionToken == "" {
			return fmt.Errorf("no session token received from signin")
		}

		if err := os.Setenv("OP_SESSION_"+account.UserUUID, sessionToken); err != nil {
			return fmt.Errorf("failed to set session token: %v", err)
		}
	} else {
		return fmt.Errorf("signin failed: %s", stderr.String())
	}

	account.SetSignInInfo(sessionToken)
	cli.Account = account

	slog.Info("connected to 1Password", "url", account.URL, "email", account.Email)
	return nil
}

// readPassword prompts the user to enter their 1Password password securely.
// It disables input echoing to ensure the password is not displayed on the screen.
// Returns the entered password as a string or an error if the password could not be read.
func readPassword() (string, error) {
	slog.Debug("prompting for 1Password password")
	fmt.Print("Enter your 1Password password: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // Add a newline after the password input
	if err != nil {
		slog.Error("failed to read password", "error", err)
		return "", err
	}
	return string(bytePassword), nil
}

// Execute runs a 1Password CLI command with the specified arguments.
// It handles both interactive and non-interactive commands, as well as
// special handling for the "signin" command.
//
// For non-interactive commands, the output is captured and returned as a byte slice.
// If an error occurs during execution, an OpCliError is returned containing
// the error and any stderr output.
//
// For the "signin" command, the function reads the user's password securely
// and pipes it into the command.
//
// For other interactive commands, the function connects the command's
// standard input, output, and error streams to the current process.
//
// Args:
//
//	args: A variadic list of strings representing the command arguments.
//
// Returns:
//
//	[]byte: The output of the command for non-interactive commands.
//	error: An error if the command fails or if there is an issue with execution.
func (cli *OpCLI) Execute(args ...string) ([]byte, error) {
	var cmdArgs []string
	if len(args) == 0 {
		return nil, fmt.Errorf("no arguments provided")
	}

	if args[0] != "signin" {
		cmdArgs = append(args, "--format=json")
	} else {
		cmdArgs = args
	}

	cmd := exec.Command(cli.Path, cmdArgs...)

	// For non-interactive commands, capture stderr and return output
	if !isInteractiveCommand(args) {
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		if err != nil {
			return nil, &OpCliError{
				Err:          err,
				StderrOutput: stderr.String(),
			}
		}
		return output, nil
	}

	// For signin command, handle password input
	if args[0] == "signin" {
		password, err := readPassword()
		if err != nil {
			return nil, fmt.Errorf("error reading password: %v", err)
		}

		command := fmt.Sprintf("%s %s", cli.Path, strings.Join(cmdArgs, " "))
		signinCmd := cli.pipePasswordCommand(password, command)
		return signinCmd.Output()
	}

	// For other interactive commands, run them directly
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// isInteractiveCommand determines if the provided command is an interactive command.
// It checks the first argument in the given list of arguments against a predefined
// set of interactive commands ("signin", "account", "user").
//
// Parameters:
//
//	args - A slice of strings representing the command-line arguments.
//
// Returns:
//
//	A boolean value:
//	  - true if the first argument matches one of the interactive commands.
//	  - false otherwise.
func isInteractiveCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}

	interactiveCommands := []string{"signin", "account", "user"}
	command := args[0]

	for _, cmd := range interactiveCommands {
		if command == cmd {
			return true
		}
	}
	return false
}

// pipePasswordCommand creates an *exec.Cmd to execute a given command with the
// specified password piped into its standard input. The password is written
// to the stdin of the command in a separate goroutine.
//
// Parameters:
//   - password: The password string to be piped into the command's stdin.
//   - command: The command to be executed, provided as a string.
//
// Returns:
//   - *exec.Cmd: The configured command object ready for execution, or nil if
//     an error occurs while creating the stdin pipe.
//
// Note:
//   - The function logs errors using slog if it fails to create the stdin pipe
//     or write the password to stdin.
func (cli *OpCLI) pipePasswordCommand(password, command string) *exec.Cmd {
	cmd := exec.Command(cli.Path, strings.Fields(command)...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		slog.Error("failed to create stdin pipe", "error", err)
		return nil
	}

	go func() {
		defer stdin.Close()
		_, err := stdin.Write([]byte(password + "\n"))
		if err != nil {
			slog.Error("failed to write password to stdin", "error", err)
		}
	}()

	return cmd
}

// ExecuteOpCommand executes a 1Password CLI command with the provided arguments.
// It ensures that account information is available and appends default arguments
// (such as the account ID) to the command before execution.
//
// Parameters:
//
//	args - A variadic list of strings representing the command-line arguments
//	       to pass to the 1Password CLI.
//
// Returns:
//
//	[]byte - The output of the executed command.
//	error  - An error if the command execution fails or if account information
//	         is missing.
//
// Errors:
//   - Returns an error if the account information is missing (Account or UserUUID is empty).
//   - Returns an error if the command execution fails, wrapping the underlying error.
//
// Example:
//
//	output, err := cli.ExecuteOpCommand("list", "items")
//	if err != nil {
//	    log.Fatalf("Command failed: %v", err)
//	}
//	fmt.Println(string(output))
func (cli *OpCLI) ExecuteOpCommand(args ...string) ([]byte, error) {
	if cli.Account == nil || cli.Account.UserUUID == "" {
		return nil, fmt.Errorf("account information is missing")
	}

	// Append --account and the account ID to the command arguments
	args = append(args, cli.getDefaultArgs()...)

	cmd := exec.Command(cli.Path, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command '%v': %w", args, err)
	}
	return output, nil
}

// containsArgument checks if a specific argument is present in a slice of strings.
// It iterates through the provided slice and returns true if the argument is found,
// otherwise it returns false.
//
// Parameters:
//   - args: A slice of strings to search through.
//   - arg: The string to look for in the slice.
//
// Returns:
//   - bool: true if the argument is found in the slice, false otherwise.
func containsArgument(args []string, arg string) bool {
	for _, a := range args {
		if a == arg {
			return true
		}
	}
	return false
}

// getDefaultArgs constructs and returns the default arguments for the OpCLI commands.
// It ensures that the account ID is included using the --account flag and appends
// the --format=json flag if it is not already present in the arguments.
//
// Returns:
//
//	[]string: A slice of strings containing the default arguments.
func (cli *OpCLI) getDefaultArgs() []string {
	var args []string

	// Append --account and the account ID to the command arguments
	args = append(args, "--account", cli.Account.UserUUID)

	// Ensure --format=json is included in the arguments
	if !containsArgument(args, "--format=json") {
		args = append(args, "--format=json")
	}
	return args
}

// updateItemWithStruct updates an existing item in 1Password using the provided Item struct.
//
// Parameters:
// - identifier: The unique identifier or name of the item to update.
// - item: The Item struct containing the updated item data.
//
// Returns:
// - error: An error object if the operation fails.
//
// This method uses the "op item edit" command to update the item.
func (cli *OpCLI) updateItemWithStruct(item Item) (*Item, error) {

	if cli.Account == nil || cli.Account.UserUUID == "" {
		return nil, fmt.Errorf("account information is missing")
	}

	args := cli.getDefaultArgs()

	// Serialize the Item struct to JSON
	jsonData, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize item to JSON: %w", err)
	}

	// Execute the "op item edit" command
	cmd := exec.Command(cli.Path, append([]string{"item", "edit", item.ID}, args...)...)
	cmd.Stdin = bytes.NewReader(jsonData)

	// Execute the "op item edit" command and capture output
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute 'op item edit': %w", err)
	}

	// Unmarshal the output into the updatedItem struct
	var updatedItem Item
	if err := json.Unmarshal(output, &updatedItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated item: %w", err)
	}

	return &updatedItem, nil
}

// CreateItem creates a new item in the 1Password vault using the "op item create" command.
// It accepts an Item object and a boolean flag indicating whether to generate a password.
//
// Parameters:
//   - item: A pointer to the Item struct representing the item to be created. The ID field
//     of the item must be empty for new items.
//   - genPassword: A boolean flag indicating whether to generate a password for the item.
//
// Returns:
//   - A pointer to the created Item struct populated with the details of the newly created item.
//   - An error if the operation fails, such as when the item ID is not empty, account information
//     is missing, JSON serialization fails, the "op item create" command fails, or the output
//     cannot be unmarshaled.
//
// Notes:
//   - The function requires the OpCLI instance to have valid account information (Account.UserUUID).
//   - The "op" CLI tool must be installed and accessible via the path specified in the OpCLI.Path field.
func (cli *OpCLI) CreateItem(item *Item, genPassword bool) (*Item, error) {

	if item.ID != "" {
		return nil, fmt.Errorf("item ID should be empty for new items")
	}

	if cli.Account == nil || cli.Account.UserUUID == "" {
		return nil, fmt.Errorf("account information is missing")
	}

	args := cli.getDefaultArgs()

	jsonData, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize item to JSON: %w", err)
	}

	var cmd *exec.Cmd
	if genPassword {
		// Generate a password if required
		cmd = exec.Command(cli.Path, append([]string{"item", "create", "--generate-password"}, args...)...)
	} else {
		cmd = exec.Command(cli.Path, append([]string{"item", "create"}, args...)...)
	}
	cmd.Stdin = bytes.NewReader(jsonData)

	// Execute the "op item create" command and capture output
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute 'op item create': %w", err)
	}

	// Unmarshal the output into the createdItem struct
	var createdItem Item
	if err := json.Unmarshal(output, &createdItem); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created item: %w", err)
	}

	return &createdItem, nil
}

// deleteItem deletes an item by its ID using the 1Password CLI.
//
// Parameters:
// - itemID: A string representing the unique identifier of the item to delete.
//
// Returns:
// - error: An error object if the operation fails.
func (cli *OpCLI) deleteItem(item Item) error {
	if item.ID == "" {
		return fmt.Errorf("item ID cannot be empty")
	}

	_, err := cli.ExecuteOpCommand("item", "delete", item.ID)
	if err != nil {
		return fmt.Errorf("failed to delete item with ID '%s': %v", item.ID, err)
	}

	return nil
}
