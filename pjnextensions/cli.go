package pjnextensions

import (
	"path"

	"github.com/spf13/cobra"
)

func Setup(extensionID string) {
	onceApp.Do(func() {
		rCmd := GetRootCmd()
		rCmd.PersistentFlags().Bool("prod", false, "Is production?")
		rCmd.PersistentFlags().StringP("root", "r", "../../", "Root Directory (default ../)")
		rCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			isProd, _ := cmd.Flags().GetBool("prod")
			rootDir := "../../"
			if isProd {
				rootDir, _ = cmd.Flags().GetString("root")
			}
			extension = &Extension{
				ID:           extensionID,
				RootDir:      rootDir,
				ExtensionDir: path.Join(extensionID, "ext"),
				DataDir:      path.Join(extensionID, "data"),
			}
			return nil
		}
	})
}
