package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shanmugharajk/gogit/internal/commit"
	"github.com/shanmugharajk/gogit/internal/file"
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
	workspace := workspace.New(cwd)
	db := storage.New(dbPath)

	// List all files in the workspace
	files, err := workspace.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list workspace files: %w", err)
	}

	entries := make([]object.Entry, 0, len(files))

	// Store each file as a blob object
	for _, filePath := range files {
		data, err := workspace.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		// Create a blob object and store it
		blob := object.NewBlob(data)
		if err := db.Store(blob); err != nil {
			return fmt.Errorf("failed to store blob for %s: %w", filePath, err)
		}

		entries = append(entries, *object.NewEntry(filePath, blob.GetOID()))
	}

	tree := object.NewTree(entries)
	if err := db.Store(tree); err != nil {
		return fmt.Errorf("failed to store tree: %w", err)
	}

	// Get author info from environment
	name := os.Getenv("GIT_AUTHOR_NAME")
	email := os.Getenv("GIT_AUTHOR_EMAIL")
	author := commit.NewAuthor(name, email, time.Now())

	// Read commit message from stdin
	messageBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read commit message: %w", err)
	}
	message := string(messageBytes)

	// Create and store commit
	commitObj := commit.NewCommit(tree.GetOID(), author, message)
	if err := db.Store(commitObj); err != nil {
		return fmt.Errorf("failed to store commit: %w", err)
	}

	// Update HEAD
	headPath := filepath.Join(gitPath, "HEAD")
	headContent := fmt.Appendf(nil, "%s\n", commitObj.GetOID())
	if err := os.WriteFile(headPath, headContent, file.ModeFile); err != nil {
		return fmt.Errorf("failed to write HEAD: %w", err)
	}

	firstLine := message
	if i := strings.IndexByte(message, '\n'); i >= 0 {
		firstLine = message[:i]
	}
	fmt.Printf("[(root-commit) %s] %s\n", commitObj.GetOID(), firstLine)

	return nil
}
