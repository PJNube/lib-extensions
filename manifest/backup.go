package manifest

const DefaultBackupResourceManifestSubject = "get.system.backup.resource-manifest"

type Backup struct {
	Enabled    bool   `json:"enabled,omitempty"`
	SkipFailed bool   `json:"skipFailed,omitempty"`
	Subject    string `json:"subject,omitempty"`
}

func (b *Backup) EffectiveSubject() string {
	if b == nil || b.Subject == "" {
		return DefaultBackupResourceManifestSubject
	}
	return b.Subject
}

type BackupResourceManifestRequest struct {
	Version string            `json:"version"`
	Tier    string            `json:"tier"`
	Roots   map[string]string `json:"roots,omitempty"`
}

type BackupResourceManifestResponse struct {
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
