package pjnextensions

import (
	"sync"

	"github.com/spf13/cobra"
)

var (
	extension   *Extension
	rootCmd     *cobra.Command
	onceApp     sync.Once
	onceRootCmd sync.Once
)

type Extension struct {
	ExtensionDir string `json:"extensionDir"`
	DataDir      string `json:"dataDir"`
	ConfigDir    string `json:"configDir"`
}

func GetExtension() *Extension {
	return extension
}

func GetRootCmd() *cobra.Command {
	onceRootCmd.Do(func() {
		rootCmd = &cobra.Command{
			Run: func(cmd *cobra.Command, args []string) {},
		}
	})
	return rootCmd
}
