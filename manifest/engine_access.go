package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
)

// SHM read buckets a CE extension can ask the control engine for.
const (
	ShmReadFull = "full"
	ShmReadNone = "none"
)

// EngineAccessKey is the extension.json key holding a CE extension's engine
// access. The control engine reads it from the extension.json next to the
// extension's manifest.json and refuses to load an extension without it.
const EngineAccessKey = "privilegedEngineAccess"

// EngineAccessKeys are the keys an EngineAccess may hold. The engine also
// refuses a manifest.json that carries any of them.
var EngineAccessKeys = []string{"shmRead", "delegatedWrite", "auditRead", "engineActions"}

// delegatedWriteCapabilities are the capabilities delegatedWrite may grant.
var delegatedWriteCapabilities = []string{
	"updateValues", "override", "callActions", "annotate", "add", "remove", "move", "copy", "edges", "userAccess",
}

// EngineAccess is what a CE extension asks the control engine for. Only
// shmRead is kept: it is the one value the platform acts on (the SHM readers
// group and the unit's --shm-bucket); the engine reads the grants itself from
// the installed extension.json, which is never rewritten from this type.
// Decoding still refuses what the engine refuses at load (control-engine
// manifest_access_parse.hpp, parseAccess) down to the capability names, so a
// package the engine would not load is refused at upload; the engine still
// checks the scopes inside each capability and the action names.
type EngineAccess struct {
	// ShmRead is ShmReadFull or ShmReadNone.
	ShmRead string `json:"shmRead"`
}

func (a *EngineAccess) UnmarshalJSON(data []byte) error {
	fields, err := jsonObject(data, EngineAccessKey)
	if err != nil {
		return err
	}
	for key := range fields {
		if !slices.Contains(EngineAccessKeys, key) {
			return fmt.Errorf("%s has an unknown key %q", EngineAccessKey, key)
		}
	}

	var access EngineAccess
	if raw, ok := fields["shmRead"]; ok {
		_ = json.Unmarshal(raw, &access.ShmRead)
	}
	if access.ShmRead != ShmReadFull && access.ShmRead != ShmReadNone {
		return fmt.Errorf("%s.shmRead must be %q or %q", EngineAccessKey, ShmReadFull, ShmReadNone)
	}

	// delegatedWrite maps a capability (e.g. "add") to its scope object. It is
	// required; an empty object grants nothing.
	raw, ok := fields["delegatedWrite"]
	if !ok {
		return fmt.Errorf("%s.delegatedWrite is required (an empty object grants nothing)", EngineAccessKey)
	}
	delegatedWrite, err := jsonObject(raw, EngineAccessKey+".delegatedWrite")
	if err != nil {
		return err
	}
	for capability, scope := range delegatedWrite {
		if !slices.Contains(delegatedWriteCapabilities, capability) {
			return fmt.Errorf("%s.delegatedWrite names an unknown capability %q", EngineAccessKey, capability)
		}
		if _, err := jsonObject(scope, EngineAccessKey+".delegatedWrite."+capability); err != nil {
			return err
		}
	}

	if raw, ok := fields["auditRead"]; ok {
		if _, err := jsonObject(raw, EngineAccessKey+".auditRead"); err != nil {
			return err
		}
	}
	// engineActions maps an engine action to its options object.
	if raw, ok := fields["engineActions"]; ok {
		engineActions, err := jsonObject(raw, EngineAccessKey+".engineActions")
		if err != nil {
			return err
		}
		for action, options := range engineActions {
			if _, err := jsonObject(options, EngineAccessKey+".engineActions."+action); err != nil {
				return err
			}
		}
	}

	*a = access
	return nil
}

// jsonObject decodes data as a JSON object (not null), naming what in errors.
func jsonObject(data json.RawMessage, what string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if trimmed := bytes.TrimSpace(data); len(trimmed) == 0 || trimmed[0] != '{' || json.Unmarshal(trimmed, &fields) != nil {
		return nil, fmt.Errorf("%s must be an object", what)
	}
	return fields, nil
}
