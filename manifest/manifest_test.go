package manifest

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPermissionsUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    Permissions
		wantErr string
	}{
		{
			name: "names with reasons",
			in:   `[{"name":"SetTimeZone","reason":"Lets the user pick the site timezone"},{"name":"ManageSsh","reason":" Remote support "}]`,
			want: Permissions{{Name: "SetTimeZone", Reason: "Lets the user pick the site timezone"}, {Name: "ManageSsh", Reason: "Remote support"}},
		},
		{name: "empty", in: `[]`, want: Permissions{}},
		{name: "plain strings", in: `["SetTimeZone"]`, wantErr: "object with"},
		{name: "not array", in: `{"name":"SetTimeZone"}`, wantErr: "cannot unmarshal"},
		{name: "lowercase name", in: `[{"name":"setTimeZone","reason":"x"}]`, wantErr: "PascalCase"},
		{name: "raw command name", in: `[{"name":"/usr/bin/date -u -s *","reason":"x"}]`, wantErr: "PascalCase"},
		{name: "missing reason", in: `[{"name":"SetTimeZone"}]`, wantErr: "reason is required"},
		{name: "blank reason", in: `[{"name":"SetTimeZone","reason":"   "}]`, wantErr: "reason is required"},
		{name: "multi-line reason", in: `[{"name":"SetTimeZone","reason":"a\nb"}]`, wantErr: "single line"},
		{name: "long reason", in: `[{"name":"SetTimeZone","reason":"` + strings.Repeat("a", 301) + `"}]`, wantErr: "at most 300"},
		{name: "duplicate", in: `[{"name":"SetTimeZone","reason":"a"},{"name":"SetTimeZone","reason":"b"}]`, wantErr: "more than once"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got Permissions
			err := json.Unmarshal([]byte(tc.in), &got)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("entry %d = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestMetadataPermissions(t *testing.T) {
	var m Metadata
	if err := json.Unmarshal([]byte(`{"name":"Platform","permissions":[{"name":"SetTimeZone","reason":"Site timezone"}]}`), &m); err != nil {
		t.Fatal(err)
	}
	if names := m.Permissions.Names(); len(names) != 1 || names[0] != "SetTimeZone" {
		t.Fatalf("Names() = %v", names)
	}
}
