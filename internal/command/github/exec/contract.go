package exec

import "github.com/spf13/cobra"

type Runner = func(*cobra.Command, []string) error
