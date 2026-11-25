package manifest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// DataCleanupKey is the extension.json key that offers an extension's data
// to the platform's disk cleaner.
const DataCleanupKey = "dataCleanup"

// DefaultDataCleanupPath is the subject, relative to the extension's own
// subjects, the disk cleaner calls when dataCleanup.path is absent.
const DefaultDataCleanupPath = "post.api.data-cleanup"

const (
	maxDataCleanupLabel       = 64
	maxDataCleanupDescription = 300
	maxDataCleanupPath        = 128
	maxDataCleanupNodePath    = 255
)

// dataCleanupPathToken is one dot-separated token of dataCleanup.path. No
// wildcards and no other characters, so the path stays inside the
// extension's own subjects.
var dataCleanupPathToken = regexp.MustCompile(`^[a-z0-9-]+$`)

// dataCleanupNodePath is a control-engine node path, e.g. "/alerts/history".
var dataCleanupNodePath = regexp.MustCompile(`^/[A-Za-z0-9_/-]*$`)

// DataCleanup offers an extension's data to the platform's disk cleaner, which
// lists it as one option. The cleaner either calls the extension at Path, a
// subject relative to the extension's own subjects, or, when CEAction is set,
// runs the dataCleanup action on a control-engine node.
type DataCleanup struct {
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	// Path is DefaultDataCleanupPath when the manifest gives none and no
	// CEAction.
	Path     string               `json:"path,omitempty"`
	CEAction *DataCleanupCEAction `json:"ceAction,omitempty"`
}

type DataCleanupCEAction struct {
	NodePath string `json:"nodePath"`
}

func (d *DataCleanup) UnmarshalJSON(data []byte) error {
	type Alias DataCleanup
	var c Alias
	if err := json.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("%s: %w", DataCleanupKey, err)
	}

	c.Label = strings.TrimSpace(c.Label)
	if c.Label == "" {
		return fmt.Errorf("%s.label is required", DataCleanupKey)
	}
	if err := dataCleanupText("label", c.Label, maxDataCleanupLabel); err != nil {
		return err
	}
	c.Description = strings.TrimSpace(c.Description)
	if err := dataCleanupText("description", c.Description, maxDataCleanupDescription); err != nil {
		return err
	}

	if c.CEAction != nil {
		if c.Path != "" {
			return fmt.Errorf("%s: path and ceAction cannot both be set", DataCleanupKey)
		}
		if len(c.CEAction.NodePath) > maxDataCleanupNodePath || !dataCleanupNodePath.MatchString(c.CEAction.NodePath) {
			return fmt.Errorf("%s.ceAction.nodePath must start with / and hold only letters, digits, _, - and /", DataCleanupKey)
		}
	} else {
		if c.Path == "" {
			c.Path = DefaultDataCleanupPath
		}
		if err := validDataCleanupPath(c.Path); err != nil {
			return err
		}
	}

	*d = DataCleanup(c)
	return nil
}

func validDataCleanupPath(path string) error {
	if len(path) > maxDataCleanupPath {
		return fmt.Errorf("%s.path must be at most %d characters", DataCleanupKey, maxDataCleanupPath)
	}
	for _, token := range strings.Split(path, ".") {
		if !dataCleanupPathToken.MatchString(token) {
			return fmt.Errorf("%s.path must be dot-separated tokens of lowercase letters, digits and -", DataCleanupKey)
		}
	}
	return nil
}

func dataCleanupText(field, value string, maxLen int) error {
	if len([]rune(value)) > maxLen {
		return fmt.Errorf("%s.%s must be at most %d characters", DataCleanupKey, field, maxLen)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("%s.%s must be a single line of text", DataCleanupKey, field)
		}
	}
	return nil
}
