package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const MetadataFileName = "extension.json"

type Dependency struct {
	Version string `json:"version"`
}

type Dependencies struct {
	Architecture string                `json:"architecture,omitempty"`
	SysVersion   string                `json:"sysVersion"`
	Platform     map[string]Dependency `json:"platform,omitempty"`
	Extension    map[string]Dependency `json:"extension,omitempty"`
}

type Resource struct {
	ID          string   `json:"id"`
	Description string   `json:"description,omitempty"`
	Items       []string `json:"items,omitempty"`
}

type DataAccess struct {
	ID         string `json:"id"`
	Permission string `json:"permission"`
}

type OpenAPIFile struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

type Metadata struct {
	Profile   string `json:"profile"`
	Vendor    string `json:"vendor"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`

	BuildUser      string              `json:"buildUser,omitempty"`
	Description    string              `json:"description,omitempty"`
	Dependencies   Dependencies        `json:"dependencies,omitempty"`
	Subjects       map[string][]string `json:"subjects,omitempty"`
	Resources      []Resource          `json:"resources,omitempty"`
	DataAccesses   []DataAccess        `json:"dataAccesses,omitempty"`
	StaticPath     string              `json:"staticPath,omitempty"`
	OpenAPISchemas []OpenAPIFile       `json:"openAPISchemas,omitempty"`
}

func GetMetadata() (*Metadata, error) {
	file, err := os.Open(MetadataFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()
	tempBytes := &bytes.Buffer{}
	_, err = file.WriteTo(tempBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	metadata := &Metadata{}
	err = json.Unmarshal(tempBytes.Bytes(), &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	return metadata, nil
}
