package builder

const (
	ExecutableName   = "main"
	ZippedFolderName = "zippedfile"
	ZipFileName      = "extension.zip"
	MetadataFileName = "extension.json"
)

type ZipMetadata struct {
	Id                    string                `json:"id"`
	Profile               string                `json:"profile"`
	Vendor                string                `json:"vendor"`
	Name                  string                `json:"name"`
	Architecture          string                `json:"architecture,omitempty"`
	Description           string                `json:"description"`
	Version               string                `json:"version"`
	SysVersion            string                `json:"sysVersion"`
	Language              string                `json:"language"`
	BuildTime             string                `json:"buildTime"`
	PlatformDependencies  *PlatformDependency   `json:"platformDependencies,omitempty"`
	ExtensionDependencies []ExtensionDependency `json:"extensionDependencies,omitempty"`
}

type PlatformDependency struct {
	BE string `json:"be,omitempty"`
	CE string `json:"ce,omitempty"`
	UI string `json:"ui,omitempty"`
}

type ExtensionDependency struct {
	Id      string `json:"id"`
	Version string `json:"version"`
}
