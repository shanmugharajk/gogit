package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shanmugharajk/gogit/internal/commit"
	"github.com/shanmugharajk/gogit/internal/object"
	"github.com/shanmugharajk/gogit/internal/refs"
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

	// Initialize workspace, storage, and refs management
	workspace := workspace.New(cwd)
	db := storage.New(dbPath)
	refsStore := refs.New(gitPath)

	parentOID, err := refsStore.ReadHead()
	if err != nil {
		return fmt.Errorf("failed to read HEAD: %w", err)
	}

	// List all files in the workspace
	files, err := workspace.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list workspace files: %w", err)
	}

	entries := make([]*object.Entry, 0, len(files))

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

		// Get file stats
		stat, err := workspace.StatFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", filePath, err)
		}

		entries = append(entries, object.NewEntry(filePath, blob.GetOID(), stat))
	}

	// Build tree hierarchy
	root := object.Build(entries)

	// Traverse and store all trees
	var storeErr error
	root.Traverse(func(tree *object.Tree) {
		if storeErr != nil {
			return // Skip if we already encountered an error
		}
		if err := db.Store(tree); err != nil {
			storeErr = fmt.Errorf("failed to store tree: %w", err)
		}
	})
	if storeErr != nil {
		return storeErr
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
	commitObj := commit.NewCommit(parentOID, root.GetOID(), author, message)
	if err := db.Store(commitObj); err != nil {
		return fmt.Errorf("failed to store commit: %w", err)
	}

	if err := refsStore.UpdateHead(commitObj.GetOID()); err != nil {
		return fmt.Errorf("failed to update HEAD: %w", err)
	}

	firstLine := message
	if i := strings.IndexByte(message, '\n'); i >= 0 {
		firstLine = message[:i]
	}
	label := ""
	if parentOID == "" {
		label = "(root-commit)"
	}
	fmt.Printf("[%s %s] %s\n", label, commitObj.GetOID(), firstLine)

	return nil
}
