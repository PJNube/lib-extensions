package privileged

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeSudo installs a script standing in for sudo that ignores its arguments
// and prints body, then exits with code.
func fakeSudo(t *testing.T, body string, code int) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "sudo")
	script := "#!/bin/sh\ncat >/dev/null\nprintf '%s' '" + body + "'\nexit " + string(rune('0'+code)) + "\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	oldSudo := SudoPath
	SudoPath = path
	t.Cleanup(func() { SudoPath = oldSudo })
}

func TestCallDecodesResult(t *testing.T) {
	fakeSudo(t, `{"ok":true,"result":{"timezone":"Asia/Kathmandu"}}`, 0)
	var out struct {
		Timezone string `json:"timezone"`
	}
	if err := Call(context.Background(), "SetTimeZone", map[string]any{"timezone": "Asia/Kathmandu"}, &out); err != nil {
		t.Fatalf("Call: %v", err)
	}
	if out.Timezone != "Asia/Kathmandu" {
		t.Fatalf("result = %q", out.Timezone)
	}
}

func TestCallReturnsBrokerError(t *testing.T) {
	fakeSudo(t, `{"ok":false,"code":"invalid_input","field":"timezone","error":"unknown timezone"}`, 2)
	err := Call(context.Background(), "SetTimeZone", map[string]any{"timezone": "Mars/Base"}, nil)
	if !IsInvalidInput(err) {
		t.Fatalf("want invalid input error, got %v", err)
	}
	if !strings.Contains(err.Error(), "timezone") {
		t.Fatalf("error should name the field: %v", err)
	}
}

func TestCallWithoutBrokerResponse(t *testing.T) {
	fakeSudo(t, ``, 1)
	err := Call(context.Background(), "SetTimeZone", nil, nil)
	if err == nil {
		t.Fatal("want error when sudo refuses")
	}
	if IsInvalidInput(err) {
		t.Fatalf("sudo refusal is not an input error: %v", err)
	}
}
