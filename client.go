package onepassword

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// opVersionPattern matches a valid 1Password CLI version string (e.g., "2.30.0").
var opVersionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// OpCLI represents the 1Password CLI executor
type OpCLI struct {
	Path             string
	accesstoken      string
	cache            itemCache
	logger           slog.Logger
	isServiceAccount bool
	Account          *Account
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

// validateVersionOutput runs "op --version" and validates the output matches
// a valid 1Password CLI semver pattern. Returns the version string on success.
func validateVersionOutput(path string) (string, error) {
	cmd := exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run '%s --version': %v", path, err)
	}

	version := strings.TrimSpace(string(output))
	if !opVersionPattern.MatchString(version) {
		return "", fmt.Errorf("unexpected version output %q: does not match expected 1Password CLI version pattern", version)
	}

	return version, nil
}

// VerifyOpExecutable verifies the authenticity of the 1Password CLI executable
// at the given path.
//
// On all platforms, it first runs "op --version" and validates that the output
// matches a valid 1Password CLI semver pattern (e.g., "2.30.0").
//
// Additional OS-specific checks:
//   - On macOS, it uses the codesign command to verify the binary is signed
//     by "Developer ID Application: AgileBits Inc."
//   - On Windows, it uses PowerShell's Get-AuthenticodeSignature to verify
//     the Authenticode signature and confirm the signer is AgileBits.
//   - On Linux, no binary-level signature verification is available. The
//     function relies on version output validation only and logs a warning.
//
// Parameters:
//   - path: The file path to the executable to be verified.
//
// Returns:
//   - An error if verification fails. Returns nil if verification succeeds.
func VerifyOpExecutable(path string) error {
	version, err := validateVersionOutput(path)
	if err != nil {
		return fmt.Errorf("version verification failed: %w", err)
	}
	slog.Debug("op version verified", "version", version)

	switch runtime.GOOS {
	case "darwin":
		if err := verifyDarwinSignature(path); err != nil {
			return err
		}
	case "windows":
		if err := verifyWindowsSignature(path); err != nil {
			return err
		}
	case "linux":
		slog.Warn("signature verification is not available for the op binary on Linux; relying on version validation only")
	default:
		slog.Warn("signature verification not supported on this platform", "os", runtime.GOOS)
	}

	return nil
}

// verifyDarwinSignature checks the macOS code signature of the binary at the
// given path and verifies it is signed by AgileBits Inc.
func verifyDarwinSignature(path string) error {
	cmd := exec.Command("codesign", "-d", "-vvv", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("macOS codesign verification failed for %s: %v\nOutput: %s", path, err, string(output))
	}

	if !strings.Contains(string(output), "Authority=Developer ID Application: AgileBits Inc.") {
		return fmt.Errorf("invalid macOS signature: binary at %s is not signed by AgileBits Inc.", path)
	}

	slog.Debug("macOS codesign verification passed", "path", path)
	return nil
}

// verifyWindowsSignature checks the Authenticode signature of the binary at
// the given path using PowerShell and verifies it is signed by AgileBits Inc.
func verifyWindowsSignature(path string) error {
	escapedPath := strings.ReplaceAll(path, "'", "''")
	psScript := fmt.Sprintf(
		`$sig = Get-AuthenticodeSignature -LiteralPath '%s'; `+
			`if ($sig.Status -ne 'Valid') { Write-Output "STATUS:$($sig.Status)"; exit 1 }; `+
			`$subject = $sig.SignerCertificate.Subject; `+
			`if ($subject -notlike '*AgileBits*') { Write-Output "SUBJECT:$subject"; exit 1 }; `+
			`Write-Output "OK:$subject"`,
		escapedPath,
	)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		if status, ok := strings.CutPrefix(outputStr, "STATUS:"); ok {
			return fmt.Errorf("Windows Authenticode verification failed: signature status is %s", status)
		}
		if subject, ok := strings.CutPrefix(outputStr, "SUBJECT:"); ok {
			return fmt.Errorf("Windows Authenticode verification failed: binary is signed by %s, not AgileBits", subject)
		}
		return fmt.Errorf("Windows signature verification failed for %s: %v\nOutput: %s", path, err, outputStr)
	}

	slog.Debug("Windows Authenticode verification passed", "path", path, "signer", outputStr)
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
	cmd.Stderr = os.Stderr
	output, err := cmd.Output() // Updated to use Output() instead of Run()
	if err != nil {
		return nil, err
	}
	return output, nil
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
