// Package privileged lets an extension run the platform actions its declared
// permissions grant (e.g. SetTimeZone).
//
// An action runs through the platform's privilege broker: a root-owned helper
// the extension's system user may invoke via sudo, once per granted action.
// The input is sent as JSON on stdin and the result comes back as JSON on
// stdout, so values never pass through a shell or the process list.
//
//	err := privileged.Call(ctx, "SetTimeZone", map[string]any{"timezone": "Asia/Kathmandu"}, nil)
package privileged

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

// BrokerPath is the platform's privilege broker. Variables so tests can point
// them at a fake.
var (
	BrokerPath = "/usr/local/bin/pjnube-priv"
	SudoPath   = "/usr/bin/sudo"
)

// Error codes returned by the broker.
const (
	CodeInvalidInput  = "invalid_input"
	CodeUnknownAction = "unknown_action"
	CodeFailed        = "failed"
)

// Error is a failure reported by the broker.
type Error struct {
	Action  string
	Code    string // one of the Code* constants
	Field   string // input field at fault, for CodeInvalidInput
	Message string
}

func (e *Error) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("privileged %s: %s: %s: %s", e.Action, e.Code, e.Field, e.Message)
	}
	return fmt.Sprintf("privileged %s: %s: %s", e.Action, e.Code, e.Message)
}

// IsInvalidInput reports whether err is the broker rejecting the input.
func IsInvalidInput(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.Code == CodeInvalidInput
}

type response struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Code   string          `json:"code,omitempty"`
	Field  string          `json:"field,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// Call runs action with input (any JSON-encodable value; nil for none) and,
// when out is non-nil, decodes the action's result into it.
func Call(ctx context.Context, action string, input, out any) error {
	if input == nil {
		input = struct{}{}
	}
	body, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("privileged %s: encode input: %w", action, err)
	}

	// -n: never prompt. A missing grant fails at once instead of hanging.
	cmd := exec.CommandContext(ctx, SudoPath, "-n", BrokerPath, action)
	cmd.Stdin = bytes.NewReader(body)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	var resp response
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		// No broker response: sudo refused (action not granted) or the
		// broker is not installed.
		if runErr != nil {
			return fmt.Errorf("privileged %s: %w: %s", action, runErr, bytes.TrimSpace(stderr.Bytes()))
		}
		return fmt.Errorf("privileged %s: invalid broker response: %w", action, err)
	}
	if !resp.OK {
		return &Error{Action: action, Code: resp.Code, Field: resp.Field, Message: resp.Error}
	}
	if out != nil && len(resp.Result) > 0 {
		if err := json.Unmarshal(resp.Result, out); err != nil {
			return fmt.Errorf("privileged %s: decode result: %w", action, err)
		}
	}
	return nil
}
