package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a new git repository",
		Long: `Initialize a new git repository in the current directory or specified path.
This command creates the necessary directory structure and files for a git repository.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runInit,
	}

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	rootPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return err
	}

	gitPath := filepath.Join(rootPath, ".git")

	dirs := []string{"objects", "refs"}
	for _, dir := range dirs {
		dirPath := filepath.Join(gitPath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Initialized empty Jit repository in %s\n", gitPath)
	return nil
}
