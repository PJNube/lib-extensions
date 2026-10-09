package manifest

type Backup struct {
	Enabled    bool   `json:"enabled,omitempty"`
	SkipFailed bool   `json:"skipFailed,omitempty"`
	Subject    string `json:"subject,omitempty"`
}

type ExtensionBackupResourceResp struct {
	Files []BackupResourceFile `json:"files"`
}

type BackupResourceFile struct {
	Root string `json:"root"`
	Path string `json:"path"`
}

type BackupResourceArchiveManifest struct {
	Version string              `json:"version"`
	Files   []BackupArchiveFile `json:"files"`
}

type BackupArchiveFile struct {
	Root        string `json:"root"`
	Path        string `json:"path"`
	ArchivePath string `json:"archivePath"`
}
