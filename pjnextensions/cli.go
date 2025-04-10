package pjnextensions

import (
	"github.com/spf13/cobra"
	"path"
)

func Setup(extensionID string) {
	onceApp.Do(func() {
		rCmd := GetRootCmd()
		rCmd.PersistentFlags().Bool("prod", false, "Is production?")
		rCmd.PersistentFlags().StringP("root", "r", "/ros/server/extensions", "Root Directory (default /ros/server/extensions)")
		rCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			isProd, _ := cmd.Flags().GetBool("prod")
			rootDir := "./"
			if !isProd {
				rootDir, _ = cmd.Flags().GetString("root")
			}
			extension = &Extension{
				ID:           extensionID,
				RootDir:      rootDir,
				ExtensionDir: path.Join(extensionID),
				DataDir:      path.Join(extensionID, "data"),
			}
			return nil
		}
	})
}
