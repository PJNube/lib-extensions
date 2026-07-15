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

type OpenAPISchema struct {
	Label string `json:"label"`
	Path  string `json:"path"`
}

// ACL permission levels for a PrivilegedPath.
const (
	ACLNone      = ""   // no ACL grant
	ACLRead      = "r"  // read-only grant
	ACLReadWrite = "rw" // read-write grant
)

// PrivilegedPath declares a host path an extension needs privileged handling
// for. ACL controls the access grant applied at install time; Backup controls
// whether the path is captured in system backups. Both default to their zero
// value (no grant, no backup), so every entry must opt in explicitly.
type PrivilegedPath struct {
	Path   string `json:"path"`
	ACL    string `json:"acl"`
	Backup bool   `json:"backup"`
}

func (p *PrivilegedPath) UnmarshalJSON(data []byte) error {
	type Alias PrivilegedPath
	if err := json.Unmarshal(data, (*Alias)(p)); err != nil {
		return err
	}
	if p.Path == "" {
		return fmt.Errorf("privilegedPaths entry: path is required")
	}
	if p.Path[0] != '/' {
		return fmt.Errorf("privilegedPaths entry %q: path must be absolute", p.Path)
	}
	switch p.ACL {
	case ACLNone, ACLRead, ACLReadWrite:
	default:
		return fmt.Errorf("privilegedPaths entry %q: invalid acl %q (must be \"\", \"r\" or \"rw\")", p.Path, p.ACL)
	}
	if p.ACL == ACLNone && !p.Backup {
		return fmt.Errorf("privilegedPaths entry %q: declares neither acl nor backup", p.Path)
	}
	return nil
}

type PrivilegedPaths []PrivilegedPath

// ACLPaths returns the entries that request an ACL grant.
func (ps PrivilegedPaths) ACLPaths() PrivilegedPaths {
	var out PrivilegedPaths
	for _, p := range ps {
		if p.ACL != ACLNone {
			out = append(out, p)
		}
	}
	return out
}

// BackupPatterns returns the path patterns of entries that request backup.
func (ps PrivilegedPaths) BackupPatterns() []string {
	var out []string
	for _, p := range ps {
		if p.Backup {
			out = append(out, p.Path)
		}
	}
	return out
}

type Metadata struct {
	Profile       string `json:"profile"`
	Vendor        string `json:"vendor"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	BuildTime     string `json:"buildTime"`
	Uninstallable bool   `json:"uninstallable,omitempty"`

	BuildUser          string              `json:"buildUser,omitempty"`
	Description        string              `json:"description,omitempty"`
	Dependencies       Dependencies        `json:"dependencies,omitempty"`
	Subjects           map[string][]string `json:"subjects,omitempty"`
	Resources          []Resource          `json:"resources,omitempty"`
	DataAccesses       []DataAccess        `json:"dataAccesses,omitempty"`
	StaticPath         string              `json:"staticPath,omitempty"`
	OpenAPISchemas     []OpenAPISchema     `json:"openAPISchemas,omitempty"`
	ReadMe             string              `json:"readMe,omitempty"`
	ChangeLog          string              `json:"changeLog,omitempty"`
	PrivilegedCommands []string            `json:"privilegedCommands,omitempty"`
	PrivilegedPaths    PrivilegedPaths     `json:"privilegedPaths,omitempty"`
	RebootOnRestore    bool                `json:"rebootOnRestore,omitempty"`
}

func (e *Metadata) UnmarshalJSON(data []byte) error {
	type Alias Metadata

	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	// Set defaults
	e.Uninstallable = true

	return json.Unmarshal(data, aux)
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
