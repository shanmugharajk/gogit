package main

import (
	"github.com/shanmugharajk/gogit/internal/commands"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gogit",
		Short: "A Go implementation of Git",
		Long: `GoGit is a comprehensive implementation of Git internals written in pure Go.
It serves as both an educational tool and a library for working with Git repositories.`,
	}

	// Add subcommands
	cmd.AddCommand(commands.NewInitCmd())
	cmd.AddCommand(commands.NewAddCmd())
	cmd.AddCommand(commands.NewCommitCmd())

	return cmd
}
