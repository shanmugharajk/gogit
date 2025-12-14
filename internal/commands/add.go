package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shanmugharajk/gogit/internal/index"
	"github.com/shanmugharajk/gogit/internal/object"
	"github.com/shanmugharajk/gogit/internal/storage"
	"github.com/shanmugharajk/gogit/internal/workspace"
	"github.com/spf13/cobra"
)

// NewAddCmd creates the add command.
func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add file contents to the index",
		Long: `Add file contents to the index (staging area).
This command reads a file from the workspace, stores it as a blob object,
and adds it to the index.`,
		Args: cobra.ExactArgs(1),
		RunE: runAdd,
	}

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Construct paths
	gitPath := filepath.Join(cwd, ".git")
	dbPath := filepath.Join(gitPath, "objects")
	indexPath := filepath.Join(gitPath, "index")

	// Initialize workspace, storage, and index
	ws := workspace.New(cwd)
	db := storage.New(dbPath)
	idx := index.New(indexPath)

	// Get the file path from arguments
	path := args[0]

	// Read file from workspace
	data, err := ws.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	// Get file stats
	stat, err := ws.StatFile(path)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", path, err)
	}

	// Create blob and store it
	blob := object.NewBlob(data)
	if err := db.Store(blob); err != nil {
		return fmt.Errorf("failed to store blob for %s: %w", path, err)
	}

	// Add entry to index
	idx.Add(path, blob.GetOID(), stat)

	// Write index updates
	if err := idx.WriteUpdates(); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}
