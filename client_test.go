package onepassword

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestOpVersionPattern_Valid(t *testing.T) {
	valid := []string{
		"2.30.0",
		"2.0.0",
		"3.1.4",
		"10.20.30",
	}
	for _, v := range valid {
		if !opVersionPattern.MatchString(v) {
			t.Errorf("expected %q to match opVersionPattern", v)
		}
	}
}

func TestOpVersionPattern_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"not-a-version",
		"2.30",
		"2.30.0-beta",
		"v2.30.0",
		"2.30.0 ",
		"abc.def.ghi",
	}
	for _, v := range invalid {
		if opVersionPattern.MatchString(v) {
			t.Errorf("expected %q to NOT match opVersionPattern", v)
		}
	}
}

func TestVerifyOpExecutable_InvalidPath(t *testing.T) {
	err := VerifyOpExecutable("/nonexistent/path/to/op")
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
}

func TestVerifyOpExecutable_RealBinary(t *testing.T) {
	opPath, err := exec.LookPath("op")
	if err != nil {
		t.Skip("op CLI not found in PATH, skipping integration test")
	}

	if err := VerifyOpExecutable(opPath); err != nil {
		t.Errorf("VerifyOpExecutable failed for real op binary: %v", err)
	}
}

func TestVerifyDarwinSignature_RealBinary(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping macOS-specific test")
	}

	opPath, err := exec.LookPath("op")
	if err != nil {
		t.Skip("op CLI not found in PATH, skipping integration test")
	}

	if err := verifyDarwinSignature(opPath); err != nil {
		t.Errorf("verifyDarwinSignature failed for real op binary: %v", err)
	}
}

func TestValidateVersionOutput_RealBinary(t *testing.T) {
	opPath, err := exec.LookPath("op")
	if err != nil {
		t.Skip("op CLI not found in PATH, skipping integration test")
	}

	version, err := validateVersionOutput(opPath)
	if err != nil {
		t.Fatalf("validateVersionOutput failed: %v", err)
	}
	if version == "" {
		t.Fatal("expected non-empty version string")
	}
	t.Logf("detected op version: %s", version)
}
