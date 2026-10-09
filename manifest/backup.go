package manifest

const DefaultBackupResourceSubject = "get.api.backup.resource"

type Backup struct {
	Enabled    bool   `json:"enabled,omitempty"`
	SkipFailed bool   `json:"skipFailed,omitempty"`
	Subject    string `json:"subject,omitempty"`
}

func (b *Backup) EffectiveSubject() string {
	if b == nil || b.Subject == "" {
		return DefaultBackupResourceSubject
	}
	return b.Subject
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
