package cmd

import (
	"fmt"

	ver "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

var (
	// populated by ldflags
	GitCommit string
	GitTag    string
	BuildDate string

	version    = "v0.0.1"
	prerelease = "beta" // blank if full release
)

func buildVersion() string {
	verStr := version
	if prerelease != "" {
		verStr = fmt.Sprintf("%s-%s", version, prerelease)
	}

	// check for git tag via ldflags
	if len(GitTag) > 0 {
		verStr = GitTag
	}

	// make sure we fail fast (panic) if bad version - this will get caught in CI tests
	ver.Must(ver.NewVersion(verStr))
	return verStr
}

var verCmd = &cobra.Command{
	Use:   "version",
	Short: "print the version",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(buildVersion())
	},
}

func init() {
	rootCmd.AddCommand(verCmd)
}
