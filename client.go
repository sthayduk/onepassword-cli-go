package onepassword

import (
	"bytes"
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
	path        string
	cache       itemCache
	logger      slog.Logger
	accesstoken string
	Account     *Account
}

// OpCliError represents an error from the 1Password CLI operations
type OpCliError struct {
	Err          error
	StderrOutput string
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

// TestOpCli checks if the 1Password CLI executable is available and functional
// by running the "--version" command.
//
// Parameters:
//   - opPath: The file path to the 1Password CLI executable.
//
// Returns:
//   - A boolean indicating whether the CLI is functional (true) or not (false).
//   - An error if the command execution fails.
//
// Example usage:
//
//	isAvailable, err := TestOpCli("/path/to/op")
//	if err != nil {
//	    log.Fatalf("Error checking 1Password CLI: %v", err)
//	}
//	if isAvailable {
//	    fmt.Println("1Password CLI is available.")
//	}
func TestOpCli(opPath string) (bool, error) {
	if err := exec.Command(opPath, "--version").Run(); err != nil {
		return false, err
	}
	return true, nil
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
// Upon successful sign-in, the session token is extracted and stored in an environment
// variable named "OP_SESSION_<UserUUID>", where <UserUUID> is the unique identifier
// of the account.
//
// Parameters:
//   - account: An Account struct containing the details of the 1Password account to sign in to.
//
// Returns:
//   - error: An error if the sign-in process fails, or nil if the sign-in is successful.
//
// Logs:
//   - Debug logs for each step of the sign-in process, including attempts and failures.
//   - Info logs upon successful connection to the 1Password account.
func (cli *OpCLI) SignIn(account *Account) error {
	slog.Debug("attempting to sign in to 1Password")

	slog.Debug("signing in to account",
		"account", account.UserUUID,
		"email", account.Email)

	signinCmd := exec.Command(cli.path, "signin", "--account", account.UserUUID, "--raw")
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
			slog.Debug("passwordless signin successful", "sessionToken", sessionToken)
			account.SetSignInInfo(sessionToken)
			cli.Account = account
		}

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

		cmd := exec.Command(cli.path, "signin", "--account", account.UserUUID, "--raw")
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

	cmd := exec.Command(cli.path, cmdArgs...)

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

		command := fmt.Sprintf("%s %s", cli.path, strings.Join(cmdArgs, " "))
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
	cmd := exec.Command(cli.path, strings.Fields(command)...)

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
