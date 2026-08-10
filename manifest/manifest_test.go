package manifest

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPrivilegedPathUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    PrivilegedPath
		wantErr string
	}{
		{
			name: "full object",
			in:   `{"path": "/etc/ntpsec/ntp.conf", "acl": "rw", "backup": true}`,
			want: PrivilegedPath{Path: "/etc/ntpsec/ntp.conf", ACL: "rw", Backup: true},
		},
		{
			name: "backup only",
			in:   `{"path": "/etc/localtime", "acl": "", "backup": true}`,
			want: PrivilegedPath{Path: "/etc/localtime", ACL: "", Backup: true},
		},
		{
			name: "read acl, backup omitted defaults false",
			in:   `{"path": "/etc/hosts", "acl": "r"}`,
			want: PrivilegedPath{Path: "/etc/hosts", ACL: "r", Backup: false},
		},
		{
			name:    "acl omitted and backup omitted declares nothing",
			in:      `{"path": "/etc/hosts"}`,
			wantErr: "declares neither acl nor backup",
		},
		{
			name:    "empty acl and backup false declares nothing",
			in:      `{"path": "/etc/hosts", "acl": "", "backup": false}`,
			wantErr: "declares neither acl nor backup",
		},
		{
			name:    "invalid acl w",
			in:      `{"path": "/etc/hosts", "acl": "w"}`,
			wantErr: "invalid acl",
		},
		{
			name:    "invalid acl rwx",
			in:      `{"path": "/etc/hosts", "acl": "rwx"}`,
			wantErr: "invalid acl",
		},
		{
			name:    "missing path",
			in:      `{"acl": "rw"}`,
			wantErr: "path is required",
		},
		{
			name:    "relative path",
			in:      `{"path": "etc/hosts", "acl": "rw"}`,
			wantErr: "must be absolute",
		},
		{
			name:    "legacy plain string rejected",
			in:      `"/etc/hosts"`,
			wantErr: "cannot unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got PrivilegedPath
			err := json.Unmarshal([]byte(tt.in), &got)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestMetadataUnmarshalPrivilegedPaths(t *testing.T) {
	in := `{
		"profile": "BE",
		"vendor": "Sensei",
		"name": "Platform",
		"version": "0.9.0",
		"privilegedPaths": [
			{"path": "/etc/ntpsec/ntp.conf", "acl": "rw", "backup": true},
			{"path": "/etc/localtime", "acl": "", "backup": true},
			{"path": "/etc/hosts", "acl": "r"}
		]
	}`
	var md Metadata
	if err := json.Unmarshal([]byte(in), &md); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !md.Uninstallable {
		t.Fatal("Uninstallable default must remain true")
	}
	if md.RebootOnRestore {
		t.Fatal("RebootOnRestore must default to false")
	}
	if len(md.PrivilegedPaths) != 3 {
		t.Fatalf("want 3 privileged paths, got %d", len(md.PrivilegedPaths))
	}

	acl := md.PrivilegedPaths.ACLPaths()
	if len(acl) != 2 || acl[0].Path != "/etc/ntpsec/ntp.conf" || acl[1].Path != "/etc/hosts" {
		t.Fatalf("ACLPaths wrong: %+v", acl)
	}
	patterns := md.PrivilegedPaths.BackupPatterns()
	if len(patterns) != 2 || patterns[0] != "/etc/ntpsec/ntp.conf" || patterns[1] != "/etc/localtime" {
		t.Fatalf("BackupPatterns wrong: %+v", patterns)
	}
}

func TestMetadataUnmarshalRebootOnRestore(t *testing.T) {
	var md Metadata
	if err := json.Unmarshal([]byte(`{"name": "x", "rebootOnRestore": true}`), &md); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !md.RebootOnRestore {
		t.Fatal("rebootOnRestore: true not parsed")
	}
}

func TestMetadataUnmarshalLegacyStringPathsRejected(t *testing.T) {
	var md Metadata
	err := json.Unmarshal([]byte(`{"name": "x", "privilegedPaths": ["/etc/hosts"]}`), &md)
	if err == nil {
		t.Fatal("legacy string-array privilegedPaths must be rejected")
	}
}

func TestPrivilegedCommandUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    PrivilegedCommand
		wantErr string
	}{
		{
			name: "command object",
			in:   `{"command": "/usr/bin/systemctl restart ntpsec.service"}`,
			want: PrivilegedCommand{Command: "/usr/bin/systemctl restart ntpsec.service"},
		},
		{
			name:    "missing command",
			in:      `{"reason": "set time"}`,
			wantErr: "command is required",
		},
		{
			name:    "empty command",
			in:      `{"command": ""}`,
			wantErr: "must not be empty",
		},
		{
			name:    "non-string command",
			in:      `{"command": 42}`,
			wantErr: "command must be a string",
		},
		{
			name:    "legacy plain string rejected",
			in:      `"/usr/bin/systemctl restart ntpsec.service"`,
			wantErr: "must be an object with a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got PrivilegedCommand
			err := json.Unmarshal([]byte(tt.in), &got)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Command != tt.want.Command {
				t.Fatalf("got command %q, want %q", got.Command, tt.want.Command)
			}
		})
	}
}

func TestPrivilegedCommandsCommandsAccessor(t *testing.T) {
	pcs := PrivilegedCommands{
		{Command: "/usr/bin/systemctl restart ntpsec.service"},
		{Command: "/usr/bin/nmcli connection up *"},
	}
	got := pcs.Commands()
	if len(got) != 2 || got[0] != "/usr/bin/systemctl restart ntpsec.service" || got[1] != "/usr/bin/nmcli connection up *" {
		t.Fatalf("Commands() wrong: %+v", got)
	}
}

func TestMetadataUnmarshalLegacyStringCommandsRejected(t *testing.T) {
	var md Metadata
	err := json.Unmarshal([]byte(`{"name": "x", "privilegedCommands": ["/usr/bin/date -u -s *"]}`), &md)
	if err == nil {
		t.Fatal("legacy string-array privilegedCommands must be rejected")
	}
}
