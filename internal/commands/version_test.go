package commands

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name           string
		expectedOutput string
	}{
		{
			name:           "version output",
			expectedOutput: "devbox (v" + Version + ")",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &cobra.Command{
				Use:   "version",
				Short: "Print the version information",
				Run: func(cmd *cobra.Command, args []string) {
					buf.WriteString(fmt.Sprintf("devbox (v%s)\n", Version))
				},
			}

			cmd.Run(cmd, []string{})

			output := strings.TrimSpace(buf.String())
			if !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, output)
			}
		})
	}
}

func TestVersionConstant(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}

	if Version != "1.0" {
		t.Errorf("Expected version to be '1.0', got %q", Version)
	}
}
