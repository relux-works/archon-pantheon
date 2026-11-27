package sessiondriver

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	ap "github.com/relux-works/archon-pantheon/pkg/sessiondriver"
)

const (
	DriverName        = "claude-stream-json"
	ObservedRelease   = "2.1.312"
	checkpointVersion = 1
	maxLine           = ap.MaxWrite
)

type Driver struct {
	mu sync.Mutex
	st state
}

type state struct {
	Version   int         `json:"version"`
	SessionID string      `json:"session_id,omitempty"`
	Activity  ap.Activity `json:"activity"`
	InputID   string      `json:"input_id,omitempty"`
	Accepted  bool        `json:"accepted,omitempty"`
	Completed bool        `json:"completed,omitempty"`
	Offset    uint64      `json:"offset"`
	Buffer    string      `json:"buffer,omitempty"`
	Closed    string      `json:"closed,omitempty"`
}

var _ ap.Driver = (*Driver)(nil)

func Descriptor() ap.Descriptor {
	return ap.Descriptor{
		Version:          ap.Version,
		ObservedRelease:  ObservedRelease,
		Name:             DriverName,
		PrimaryTransport: "pipe",
		DeclaredFallback: "pty",
		Capabilities: ap.Capabilities{
			PreallocatedIdentity: true,
			NextTurn:             true,
			Activity:             true,
			IdleFence:            true,
		},
	}
}

func (d *Driver) Descriptor() ap.Descriptor { return Descriptor() }

func OpenDriver(cfg ap.Config, saved []byte) (*Driver, uint64, error) {
	if (cfg.NativeRelease != "" && cfg.NativeRelease != ObservedRelease) || !Descriptor().SupportsLifecycle(cfg) {
		return nil, 0, ap.ErrUnsupported
	}
	if len(saved) == 0 {
		act := ap.Unknown
		if cfg.NativeSession != "" {
			act = ap.Idle
		}
		return &Driver{
			st: state{
				Version:   checkpointVersion,
				SessionID: cfg.NativeSession,
				Activity:  act,
			},
		}, 0, nil
	}
	st, err := decodeState(saved)
	if err != nil {
		return nil, 0, err
	}
	if cfg.NativeSession != "" && st.SessionID != "" && st.SessionID != cfg.NativeSession {
		return nil, 0, errors.New("driver_checkpoint_refused")
	}
	if st.SessionID == "" && cfg.NativeSession != "" {
		st.SessionID = cfg.NativeSession
	}
	return &Driver{st: st}, st.Offset, nil
}

func (d *Driver) Snapshot() ap.Transition {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.snapshot()
}

func (d *Driver) snapshot() ap.Transition {
	cp, _ := encodeState(d.st)
	return ap.Transition{
		View: ap.View{
			InputID:       d.st.InputID,
			Activity:      d.st.Activity,
			NativeSession: d.st.SessionID,
		},
		Checkpoint: cp,
	}
}

func (d *Driver) Begin() (ap.Transition, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.finish(nil, nil)
}

func (d *Driver) Prepare(in ap.Input) (ap.Transition, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !Descriptor().Capabilities.Supports(in.Mode) {
		return d.snapshot(), ap.ErrUnsupported
	}
	if !validIdentifier(in.ID) || !validIdentifier(in.Reference) || in.ExpectedTurn != "" {
		return d.snapshot(), errors.New("invalid_notice_framing")
	}
	if in.ID == d.st.InputID {
		return d.snapshot(), errors.New("duplicate_protocol_input")
	}
	if d.st.Closed != "" || d.st.Activity != ap.Idle || d.st.SessionID == "" {
		return d.snapshot(), errors.New("driver_input_closed")
	}
	msg := userEnvelope{
		Type:      "user",
		SessionID: d.st.SessionID,
		Message: userMessage{
			Role:    "user",
			Content: "Notice reference: " + in.Reference + ". For this synthetic notice, reply ok without tools.",
		},
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		return d.snapshot(), err
	}
	raw = append(raw, '\n')
	prev := d.st
	d.st.InputID = in.ID
	d.st.Accepted = false
	d.st.Completed = false
	d.st.Activity = ap.Busy
	sum := sha256.Sum256([]byte(in.ID))
	tx, err := d.finish([]ap.Action{{
		Key:     "notice-" + hex.EncodeToString(sum[:]),
		InputID: in.ID,
		Bytes:   raw,
	}}, nil)
	if err != nil {
		d.st = prev
		return d.snapshot(), err
	}
	return tx, nil
}

func (d *Driver) Observe(chunk ap.Chunk) (ap.Transition, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	prev := d.st
	if chunk.Offset != d.st.Offset {
		return d.snapshot(), errors.New("stream_offset_mismatch")
	}
	d.st.Offset += uint64(len(chunk.Bytes))
	if chunk.EOF {
		d.st.Closed = "eof"
		d.st.Activity = ap.Unknown
		return d.finish(nil, nil)
	}
	lines, ok := d.consumeLines(chunk.Bytes)
	if !ok {
		d.st.Closed = "frame_overflow"
		d.st.Activity = ap.Unknown
		return d.finish(nil, nil)
	}
	var facts []ap.Fact
	for _, line := range lines {
		facts = append(facts, d.observeLine(line, chunk.AtTail)...)
	}
	tx, err := d.finish(nil, facts)
	if err != nil {
		d.st = prev
		return d.snapshot(), err
	}
	return tx, nil
}

func (d *Driver) consumeLines(chunk []byte) ([][]byte, bool) {
	combined := d.st.Buffer + string(chunk)
	var lines [][]byte
	for {
		idx := strings.IndexByte(combined, '\n')
		if idx < 0 {
			break
		}
		line := strings.TrimSuffix(combined[:idx], "\r")
		combined = combined[idx+1:]
		if len(line) > maxLine {
			d.st.Buffer = ""
			return nil, false
		}
		if strings.TrimSpace(line) != "" {
			lines = append(lines, []byte(line))
		}
	}
	if len(combined) > maxLine {
		d.st.Buffer = ""
		return nil, false
	}
	d.st.Buffer = combined
	return lines, true
}

func (d *Driver) observeLine(line []byte, atTail bool) []ap.Fact {
	var env inboundEnvelope
	if err := json.Unmarshal(line, &env); err != nil || env.Type == "" {
		d.st.Activity = ap.Unknown
		d.st.Closed = "invalid_json"
		return nil
	}
	if env.SessionID != "" {
		if d.st.SessionID == "" && validIdentifier(env.SessionID) {
			d.st.SessionID = env.SessionID
		} else if env.SessionID != d.st.SessionID {
			d.st.Activity = ap.Unknown
			d.st.Closed = "session_mismatch"
			return nil
		}
	}
	var facts []ap.Fact
	switch env.Type {
	case "system":
		if d.st.Activity == ap.Unknown && d.st.SessionID != "" && d.st.Closed == "" && atTail {
			d.st.Activity = ap.Idle
		}
	case "assistant":
		if d.st.InputID != "" && d.st.SessionID != "" && !d.st.Accepted && atTail {
			d.st.Accepted = true
			d.st.Activity = ap.Busy
			facts = append(facts, ap.Fact{
				Kind:          ap.InputAccepted,
				InputID:       d.st.InputID,
				NativeSession: d.st.SessionID,
			})
		}
	case "result":
		if d.st.InputID == "" || d.st.SessionID == "" || !atTail {
			d.st.Activity = ap.Unknown
			break
		}
		if env.IsError || env.Subtype == "error" {
			d.st.Activity = ap.Unknown
			facts = append(facts, ap.Fact{
				Kind:          ap.TurnFailed,
				InputID:       d.st.InputID,
				NativeSession: d.st.SessionID,
			})
			break
		}
		if !d.st.Completed {
			d.st.Completed = true
			d.st.Activity = ap.Idle
			facts = append(facts, ap.Fact{
				Kind:          ap.TurnCompleted,
				InputID:       d.st.InputID,
				NativeSession: d.st.SessionID,
			})
		}
	}
	return facts
}

func (d *Driver) finish(actions []ap.Action, facts []ap.Fact) (ap.Transition, error) {
	tx := d.snapshot()
	tx.Actions = actions
	tx.Facts = facts
	if err := ap.Validate(Descriptor(), tx); err != nil {
		d.st.Activity = ap.Unknown
		d.st.Closed = "validate_failed"
		return d.snapshot(), errors.New("driver_transition_refused")
	}
	return tx, nil
}

func encodeState(s state) (json.RawMessage, error) {
	s.Version = checkpointVersion
	raw, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	if len(raw) > ap.MaxCheckpoint {
		return nil, errors.New("driver_checkpoint_refused")
	}
	return raw, nil
}

func decodeState(raw []byte) (state, error) {
	if len(raw) > ap.MaxCheckpoint || !json.Valid(raw) {
		return state{}, errors.New("driver_checkpoint_refused")
	}
	var st state
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&st); err != nil {
		return state{}, fmt.Errorf("driver_checkpoint_refused: %w", err)
	}
	if st.Version != checkpointVersion {
		return state{}, errors.New("driver_checkpoint_refused")
	}
	switch st.Activity {
	case ap.Unknown, ap.Idle, ap.Busy, ap.Draft, ap.Approval, ap.SecurityStop:
	default:
		return state{}, errors.New("driver_checkpoint_refused")
	}
	return st, nil
}

func validIdentifier(s string) bool {
	if len(s) == 0 || len(s) > 128 {
		return false
	}
	for _, r := range s {
		if r < 33 || r > 126 {
			return false
		}
	}
	return true
}

type userEnvelope struct {
	Type      string      `json:"type"`
	SessionID string      `json:"session_id,omitempty"`
	Message   userMessage `json:"message"`
}

type userMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type inboundEnvelope struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}
