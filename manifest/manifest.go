package manifest

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
