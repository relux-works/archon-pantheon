// Package fixtureadapters supplies two synthetic protocols to exercise plugins.
// Neither adapter implements a real agent harness.
package fixtureadapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/relux-works/curator-runtimes/pkg/sessiondriver"
)

const maxFrame = 64 << 10

type JSON struct{}

func (JSON) Name() string                           { return "json-fixture" }
func (JSON) InputState() sessiondriver.InputState   { return sessiondriver.InputIdle }

// Gated supplies explicit synthetic input evidence without parsing a real harness.
type Gated struct {
	sessiondriver.NoticeAdapter
	State sessiondriver.InputState
}

func (g Gated) InputState() sessiondriver.InputState { return g.State }

type Envelope struct {
	Kind      string                        `json:"kind"`
	Notice    sessiondriver.Notice          `json:"notice"`
	Selection sessiondriver.NoticeSelection `json:"selection"`
}

func DecodeLine(reader io.Reader, value any) error {
	data, err := io.ReadAll(io.LimitReader(reader, maxFrame+1))
	if err != nil {
		return err
	}
	if len(data) > maxFrame {
		return errors.New("frame_too_large")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return errors.New("invalid_frame")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("invalid_frame")
	}
	return nil
}

func WriteLine(writer io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(data)+1 > maxFrame {
		return errors.New("frame_too_large")
	}
	data = append(data, '\n')
	n, err := writer.Write(data)
	if err == nil && n != len(data) {
		return io.ErrShortWrite
	}
	return err
}

func (JSON) Encode(n sessiondriver.Notice, s sessiondriver.NoticeSelection) ([]byte, error) {
	data, err := json.Marshal(Envelope{Kind: "notice", Notice: n, Selection: s})
	return append(data, '\n'), err
}

func (JSON) Applied(output []byte, n sessiondriver.Notice, s sessiondriver.NoticeSelection) (bool, error) {
	lines := bytes.Split(output, []byte{'\n'})
	for _, line := range lines[:len(lines)-1] {
		var envelope Envelope
		if err := DecodeLine(bytes.NewReader(line), &envelope); err != nil || envelope.Kind != "applied" {
			return false, errors.New("invalid_fixture_ack")
		}
		if envelope.Notice == n && envelope.Selection == s {
			return true, nil
		}
	}
	return false, nil
}

// Lines uses a distinct wire encoding. Adding it requires no core changes.
type Lines struct{}

func (Lines) Name() string                         { return "line-fixture" }
func (Lines) InputState() sessiondriver.InputState { return sessiondriver.InputIdle }

func (Lines) Encode(n sessiondriver.Notice, s sessiondriver.NoticeSelection) ([]byte, error) {
	fields := []string{n.ID, n.Reference, s.Vendor, s.Model, s.Reasoning}
	for _, field := range fields {
		if strings.ContainsAny(field, "\t\r\n") {
			return nil, errors.New("invalid_fixture_field")
		}
	}
	return []byte(strings.Join(fields, "\t") + "\n"), nil
}

func (a Lines) Applied(output []byte, n sessiondriver.Notice, s sessiondriver.NoticeSelection) (bool, error) {
	expected, err := a.Encode(n, s)
	if err != nil {
		return false, err
	}
	lines := bytes.Split(output, []byte{'\n'})
	for _, line := range lines[:len(lines)-1] {
		if !bytes.HasPrefix(line, []byte("applied\t")) || bytes.Count(line, []byte{'\t'}) != 5 {
			return false, errors.New("invalid_fixture_ack")
		}
		if bytes.Equal(bytes.TrimPrefix(line, []byte("applied\t")), bytes.TrimSuffix(expected, []byte{'\n'})) {
			return true, nil
		}
	}
	return false, nil
}
