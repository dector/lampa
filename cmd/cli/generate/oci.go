package generate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dector/lampa/internal/templates/containerfile"
	"github.com/urfave/cli/v3"
)

const (
	OptToDir = "to-dir"
)

func CreateOciCommand() *cli.Command {
	return &cli.Command{
		Name:  "oci",
		Usage: "generate Containerfile",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  OptToDir,
				Usage: "directory where to put Containerfile",
				Value: ".",
			},
		},
		Action: ActionCmdOci,
	}
}

func ActionCmdOci(ctx context.Context, cmd *cli.Command) error {
	toDir := cmd.String(OptToDir)

	// Ensure directory exists
	if err := os.MkdirAll(toDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", toDir, err)
	}

	// Generate Containerfile content using template
	content, err := generateContainerfileContent()
	if err != nil {
		return fmt.Errorf("failed to generate Containerfile content: %v", err)
	}

	// Write to file
	filePath := filepath.Join(toDir, "Containerfile")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write Containerfile: %v", err)
	}

	fmt.Printf("Containerfile created at %s\n", filePath)
	return nil
}

func generateContainerfileContent() (string, error) {
	return containerfile.GenerateContainerfile()
}
