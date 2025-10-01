package pjnextensions

import (
	"path"

	"github.com/spf13/cobra"
)

func Setup(extensionID string) {
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
				ID:           extensionID,
				ExtensionDir: extensionDir,
				DataDir:      path.Join(extensionDir, "data"),
				ConfigDir:    path.Join(extensionDir, "config"),
			}
			return nil
		}
	})
}
