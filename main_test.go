package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandHelpIncludesGenerateJWTTokenCommand(t *testing.T) {
	usage := executeHelp(t, newRootCommand())
	for _, want := range []string{
		generateJWTTokenCommand,
		"--config-dir",
		"生成 JWT 私钥和公钥文件",
	} {
		if !strings.Contains(usage, want) {
			t.Fatalf("expected main usage to contain %q, got %q", want, usage)
		}
	}
}

func TestGenerateJWTTokenCommandHelpIncludesForceFlag(t *testing.T) {
	usage := executeHelp(t, newGenerateJWTTokenCommand())

	for _, want := range []string{
		generateJWTTokenCommand,
		"--force",
		"private.pem",
		"public.pem",
	} {
		if !strings.Contains(usage, want) {
			t.Fatalf("expected subcommand usage to contain %q, got %q", want, usage)
		}
	}
}

func TestGenerateJWTTokenCommandRejectsTooManyArgs(t *testing.T) {
	cmd := newGenerateJWTTokenCommand()
	cmd.SetArgs([]string{"a", "b"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for too many args")
	}

	if !strings.Contains(err.Error(), "accepts at most 1 arg") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func executeHelp(t *testing.T, cmd *cobra.Command) string {
	t.Helper()

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help failed: %v", err)
	}

	return out.String()
}
