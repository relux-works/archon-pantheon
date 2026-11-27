package sessiondriver

import (
	"testing"

	ap "github.com/relux-works/archon-pantheon/pkg/sessiondriver"
)

func TestClaudeStreamJSONLifecycleAndCheckpointRestore(t *testing.T) {
	cfg := ap.Config{
		Lifecycle:     ap.Create,
		NativeSession: "session-a",
		NativeRelease: ObservedRelease,
	}
	d, offset, err := OpenDriver(cfg, nil)
	if err != nil || offset != 0 {
		t.Fatalf("OpenDriver: err=%v offset=%d", err, offset)
	}
	tx, err := d.Begin()
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}
	if tx.View.Activity != ap.Idle || tx.View.NativeSession != "session-a" {
		t.Fatalf("unexpected initial view: %+v", tx.View)
	}

	tx, err = d.Prepare(ap.Input{ID: "notice-a", Mode: ap.NextTurn, Reference: "synthetic-reference"})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if len(tx.Actions) != 1 || tx.View.Activity != ap.Busy {
		t.Fatalf("unexpected prepare transition: %+v", tx)
	}

	restored, offset, err := OpenDriver(cfg, tx.Checkpoint)
	if err != nil {
		t.Fatalf("OpenDriver restored: %v", err)
	}

	assistantLine := []byte(`{"type":"assistant","session_id":"session-a","message":{"role":"assistant","content":[]}}` + "\n")
	tx, err = restored.Observe(ap.Chunk{
		Offset: offset,
		Bytes:  assistantLine,
		AtTail: true,
	})
	if err != nil {
		t.Fatalf("Observe assistant: %v", err)
	}
	offset += uint64(len(assistantLine))
	var sawAccepted bool
	for _, f := range tx.Facts {
		if f.Kind == ap.InputAccepted && f.InputID == "notice-a" {
			sawAccepted = true
		}
	}
	if !sawAccepted {
		t.Fatalf("expected InputAccepted fact, got %+v", tx.Facts)
	}

	resultLine := []byte(`{"type":"result","subtype":"success","is_error":false,"session_id":"session-a","result":"ok"}` + "\n")
	tx, err = restored.Observe(ap.Chunk{
		Offset: offset,
		Bytes:  resultLine,
		AtTail: true,
	})
	if err != nil {
		t.Fatalf("Observe result: %v", err)
	}
	if tx.View.Activity != ap.Idle {
		t.Fatalf("expected Idle after result, got %+v", tx.View)
	}
	var sawCompleted bool
	for _, f := range tx.Facts {
		if f.Kind == ap.TurnCompleted && f.InputID == "notice-a" && f.NativeID == "" {
			sawCompleted = true
		}
	}
	if !sawCompleted {
		t.Fatalf("expected TurnCompleted without fabricated NativeID, got %+v", tx.Facts)
	}
}
