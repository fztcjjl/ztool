package cmd

import (
	"log"

	"github.com/fztcjjl/ztool/config"
	"github.com/fztcjjl/ztool/core"
	"github.com/spf13/cobra"
)

var reverseCmd = &cobra.Command{
	Use:     "reverse",
	Short:   "Reverse a db to codes",
	Example: "ztool reverse",
	PreRun: func(cmd *cobra.Command, args []string) {
		setup()
	},
	Run: func(cmd *cobra.Command, args []string) {
		reverse()
	},
}

var dest string

func init() {
	reverseCmd.Flags().StringVarP(&dest, "dest", "d", "model", "dest directory")

	rootCmd.AddCommand(reverseCmd)
}

func setup() {
	config.Init()
}

func reverse() {
	if err := core.Generate(dest); err != nil {
		log.Println(err)
	}
}
