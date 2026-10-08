package manifest

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEngineAccess(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{name: "minimal", in: `{"shmRead":"none","delegatedWrite":{}}`},
		{name: "full grant", in: `{"shmRead":"full","delegatedWrite":{"add":{"types":"any","location":"anywhere"},"updateValues":{}},"auditRead":{},"engineActions":{"restart":{}}}`},
		{name: "not an object", in: `"full"`, wantErr: "privilegedEngineAccess must be an object"},
		{name: "array", in: `[]`, wantErr: "privilegedEngineAccess must be an object"},
		{name: "unknown key", in: `{"shmRead":"none","delegatedWrite":{},"rootShell":true}`, wantErr: `unknown key "rootShell"`},
		{name: "missing shmRead", in: `{"delegatedWrite":{}}`, wantErr: "shmRead must be"},
		{name: "bad shmRead", in: `{"shmRead":"some","delegatedWrite":{}}`, wantErr: "shmRead must be"},
		{name: "shmRead not a string", in: `{"shmRead":true,"delegatedWrite":{}}`, wantErr: "shmRead must be"},
		{name: "missing delegatedWrite", in: `{"shmRead":"full"}`, wantErr: "delegatedWrite is required"},
		{name: "null delegatedWrite", in: `{"shmRead":"full","delegatedWrite":null}`, wantErr: "delegatedWrite must be an object"},
		{name: "unknown capability", in: `{"shmRead":"none","delegatedWrite":{"format":{}}}`, wantErr: `unknown capability "format"`},
		{name: "scope not an object", in: `{"shmRead":"none","delegatedWrite":{"add":"anywhere"}}`, wantErr: "delegatedWrite.add must be an object"},
		{name: "auditRead not an object", in: `{"shmRead":"none","delegatedWrite":{},"auditRead":true}`, wantErr: "auditRead must be an object"},
		{name: "engineActions not an object", in: `{"shmRead":"none","delegatedWrite":{},"engineActions":["restart"]}`, wantErr: "engineActions must be an object"},
		{name: "engine action not an object", in: `{"shmRead":"none","delegatedWrite":{},"engineActions":{"restart":true}}`, wantErr: "engineActions.restart must be an object"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var a EngineAccess
			err := json.Unmarshal([]byte(tc.in), &a)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestMetadataEngineAccess(t *testing.T) {
	var ce Metadata
	if err := json.Unmarshal([]byte(`{"profile":"CE","vendor":"acme","name":"widget","version":"1.0.0","privilegedEngineAccess":{"shmRead":"full","delegatedWrite":{"add":{}}}}`), &ce); err != nil {
		t.Fatal(err)
	}
	if ce.PrivilegedEngineAccess == nil || ce.PrivilegedEngineAccess.ShmRead != ShmReadFull {
		t.Fatalf("engine access not decoded: %+v", ce.PrivilegedEngineAccess)
	}

	var be Metadata
	if err := json.Unmarshal([]byte(`{"profile":"BE","vendor":"acme","name":"svc","version":"1.0.0"}`), &be); err != nil {
		t.Fatal(err)
	}
	if be.PrivilegedEngineAccess != nil {
		t.Fatal("a manifest without engine access must decode to nil")
	}
	if out, _ := json.Marshal(be); strings.Contains(string(out), EngineAccessKey) {
		t.Fatalf("absent engine access must not be written: %s", out)
	}

	var bad Metadata
	err := json.Unmarshal([]byte(`{"profile":"CE","vendor":"acme","name":"widget","privilegedEngineAccess":{"shmRead":"full"}}`), &bad)
	if err == nil || !strings.Contains(err.Error(), "delegatedWrite is required") {
		t.Fatalf("an invalid declaration must fail the manifest, got %v", err)
	}
}
