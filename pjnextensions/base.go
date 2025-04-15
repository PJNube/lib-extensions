package pjnextensions

import (
	"github.com/spf13/cobra"
	"sync"
)

var (
	extension   *Extension
	rootCmd     *cobra.Command
	onceApp     sync.Once
	onceRootCmd sync.Once
)

type Extension struct {
	ID           string `json:"Id"`
	RootDir      string `json:"rootDir"`
	ExtensionDir string `json:"extensionDir"`
	DataDir      string `json:"dataDir"`
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
