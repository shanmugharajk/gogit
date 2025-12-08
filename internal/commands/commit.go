package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shanmugharajk/gogit/internal/object"
	"github.com/shanmugharajk/gogit/internal/storage"
	"github.com/shanmugharajk/gogit/internal/workspace"
	"github.com/spf13/cobra"
)

// NewCommitCmd creates the commit command.
func NewCommitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Record changes to the repository",
		Long: `Create a commit that records the current state of the workspace.
This command reads all files in the workspace and stores them as blob objects in the database.`,
		RunE: runCommit,
	}

	return cmd
}

// runCommit executes the commit command.
func runCommit(cmd *cobra.Command, args []string) error {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Construct paths
	gitPath := filepath.Join(cwd, ".git")
	dbPath := filepath.Join(gitPath, "objects")

	// Initialize workspace and database
	ws := workspace.New(cwd)
	db := storage.New(dbPath)

	// List all files in the workspace
	files, err := ws.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list workspace files: %w", err)
	}

	// Store each file as a blob object
	for _, filePath := range files {
		data, err := ws.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Create a blob object and store it
		blob := object.NewBlob(data)
		if err := db.Store(blob); err != nil {
			return fmt.Errorf("failed to store blob for %s: %w", filePath, err)
		}
	}

	return nil
}
