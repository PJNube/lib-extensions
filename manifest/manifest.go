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

// ACL permission levels for a PrivilegedPath.
const (
	ACLNone      = ""   // no ACL grant
	ACLRead      = "r"  // read-only grant
	ACLReadWrite = "rw" // read-write grant
)

// PrivilegedPath declares a host path an extension needs privileged handling
// for.
//
// Deprecated: declare Permissions instead. The platform owns which files a
// permission touches and whether they are backed up. ACL controls the access grant applied at install time; Backup controls
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

// PrivilegedCommand declares a platform command an extension needs to run
// with elevated privileges.
//
// Deprecated: declare Permissions instead and call the granted actions with
// the privileged package. Command must be exactly one of the commands in
// the platform's allowlist; the server resolves the human-readable
// description from that allowlist and never reads it from the manifest.
type PrivilegedCommand struct {
	Command string `json:"command"`
}

func (pc *PrivilegedCommand) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("privilegedCommands entry must be an object with a \"command\" key")
	}
	raw, ok := fields["command"]
	if !ok {
		return fmt.Errorf("privilegedCommands entry: command is required")
	}
	if err := json.Unmarshal(raw, &pc.Command); err != nil {
		return fmt.Errorf("privilegedCommands entry: command must be a string")
	}
	if pc.Command == "" {
		return fmt.Errorf("privilegedCommands entry: command must not be empty")
	}
	return nil
}

type PrivilegedCommands []PrivilegedCommand

// Commands returns the command strings in declaration order.
func (ps PrivilegedCommands) Commands() []string {
	out := make([]string, 0, len(ps))
	for _, pc := range ps {
		out = append(out, pc.Command)
	}
	return out
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
	Permissions        Permissions         `json:"permissions,omitempty"`
	PrivilegedCommands PrivilegedCommands  `json:"privilegedCommands,omitempty"` // Deprecated: use Permissions.
	PrivilegedPaths    PrivilegedPaths     `json:"privilegedPaths,omitempty"`    // Deprecated: use Permissions.
	RebootOnRestore    bool                `json:"rebootOnRestore,omitempty"`    // Deprecated: derived from Permissions.
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
