package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/ctrl-vfr/persona/internal/ui"
	"github.com/spf13/cobra"
)

// Version information - set at build time
var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
	gitBranch = "unknown"
)

var (
	outputJSON  bool
	outputPlain bool
)

// VersionInfo contains version and build information
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

func getVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   version,
		BuildTime: buildTime,
		GitCommit: gitCommit,
		GitBranch: gitBranch,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func getVersionString() string {
	if version == "dev" {
		return fmt.Sprintf("persona %s-%s (%s %s)", version, gitCommit, runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Sprintf("persona %s (%s %s)", version, runtime.GOOS, runtime.GOARCH)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  "Display persona version along with build and platform information",
	Run: func(cmd *cobra.Command, args []string) {
		info := getVersionInfo()

		if outputJSON {
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		if outputPlain {
			fmt.Println(info.Version)
			return
		}

		// Default: full formatted output
		fmt.Println(ui.TitleStyle.Render("Version"))
		fmt.Println(ui.ContentStyle.Render(fmt.Sprintf("Version: %s", info.Version)))
		fmt.Println(ui.ContentStyle.Render(fmt.Sprintf("Build time: %s", info.BuildTime)))
		fmt.Println(ui.ContentStyle.Render(fmt.Sprintf("Git commit: %s", info.GitCommit)))
		fmt.Println(ui.ContentStyle.Render(fmt.Sprintf("Git branch: %s", info.GitBranch)))
		fmt.Println(ui.ContentStyle.Render(fmt.Sprintf("Go version: %s", info.GoVersion)))
		fmt.Println(ui.ContentStyle.Render(fmt.Sprintf("Platform: %s\n", info.Platform)))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&outputJSON, "json", false, "Display information in JSON format")
	versionCmd.Flags().BoolVar(&outputPlain, "plain", false, "Display simple version only")
}
