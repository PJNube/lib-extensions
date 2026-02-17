package pjnextensions

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
)

func Setup(version string) {
	onceApp.Do(func() {
		rCmd := GetRootCmd()
		rCmd.PersistentFlags().Bool("prod", false, "Is production?")
		rCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			isProd, _ := cmd.Flags().GetBool("prod")
			extensionDir := "./"
			if isProd {
				extensionDir = "../"
			}
			extension = &Extension{
				ExtensionDir: extensionDir,
				DataDir:      path.Join(extensionDir, "data"),
				ConfigDir:    path.Join(extensionDir, "config"),
			}

			return nil
		}
		SetVersion(version)
	})
}

func SetVersion(version string) {
	rCmd := GetRootCmd()
	// Set version
	if version == "" {
		version = "dev"
	}
	rCmd.Version = version

	// Enhanced version template with build information
	versionTemplate := fmt.Sprintf(`%s
`, version)
	rCmd.SetVersionTemplate(versionTemplate)
}
