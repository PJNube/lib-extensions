package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
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

// permissionNamePattern matches a permission name such as "SetTimeZone".
var permissionNamePattern = regexp.MustCompile(`^[A-Z][A-Za-z0-9]{1,63}$`)

// maxPermissionReason caps a permission's reason, shown to the user at install.
const maxPermissionReason = 300

// Permission is a platform permission an extension requests, by name (e.g.
// "SetTimeZone"), with the reason it needs it. The platform owns what the name
// grants: the privileged actions it unlocks, its risk level, and which system
// files it backs up. The reason is shown to the user next to the platform's
// own description. An unknown name fails the install.
type Permission struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

func (p *Permission) UnmarshalJSON(data []byte) error {
	type Alias Permission
	if err := json.Unmarshal(data, (*Alias)(p)); err != nil {
		return fmt.Errorf("permissions entry must be an object with \"name\" and \"reason\"")
	}
	if !permissionNamePattern.MatchString(p.Name) {
		return fmt.Errorf("permissions entry %q: name must be PascalCase letters and digits", p.Name)
	}
	p.Reason = strings.TrimSpace(p.Reason)
	if p.Reason == "" {
		return fmt.Errorf("permissions entry %q: reason is required", p.Name)
	}
	if len([]rune(p.Reason)) > maxPermissionReason {
		return fmt.Errorf("permissions entry %q: reason must be at most %d characters", p.Name, maxPermissionReason)
	}
	for _, r := range p.Reason {
		if unicode.IsControl(r) {
			return fmt.Errorf("permissions entry %q: reason must be a single line of text", p.Name)
		}
	}
	return nil
}

type Permissions []Permission

func (ps *Permissions) UnmarshalJSON(data []byte) error {
	var entries []Permission
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(entries))
	for _, p := range entries {
		if _, ok := seen[p.Name]; ok {
			return fmt.Errorf("permissions entry %q: declared more than once", p.Name)
		}
		seen[p.Name] = struct{}{}
	}
	*ps = entries
	return nil
}

// Names returns the requested permission names in declaration order.
func (ps Permissions) Names() []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.Name)
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

	BuildUser      string              `json:"buildUser,omitempty"`
	Description    string              `json:"description,omitempty"`
	Dependencies   Dependencies        `json:"dependencies,omitempty"`
	Subjects       map[string][]string `json:"subjects,omitempty"`
	Resources      []Resource          `json:"resources,omitempty"`
	DataAccesses   []DataAccess        `json:"dataAccesses,omitempty"`
	StaticPath     string              `json:"staticPath,omitempty"`
	OpenAPISchemas []OpenAPISchema     `json:"openAPISchemas,omitempty"`
	ReadMe         string              `json:"readMe,omitempty"`
	ChangeLog      string              `json:"changeLog,omitempty"`
	Permissions    Permissions         `json:"permissions,omitempty"`
	// PrivilegedEngineAccess is a CE extension's control-engine access; nil
	// when the manifest declares none (BE and UI extensions).
	PrivilegedEngineAccess *EngineAccess `json:"privilegedEngineAccess,omitempty"`
	Backup                 *Backup       `json:"backup,omitempty"`
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
