package io_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	. "go.octolab.org/toolset/maintainer/internal/pkg/io"
)

func TestPacker(t *testing.T) {
	packer := NewPacker()

	packer.Register(
		func(w io.Writer) Encoder { return json.NewEncoder(w) },
		func(r io.Reader) Decoder { return json.NewDecoder(r) },
		".json",
	)

	t.Run("supported format", func(t *testing.T) {
		f := &file{Buffer: bytes.NewBuffer(nil), name: "test.json"}
		assert.NoError(t, packer.Pack(f, map[string]string{"foo": "bar"}))
		assert.JSONEq(t, `{"foo":"bar"}`, f.String())

		var data map[string]string
		assert.NoError(t, packer.Unpack(f, &data))
		assert.Equal(t, map[string]string{"foo": "bar"}, data)
	})

	t.Run("unsupported format", func(t *testing.T) {
		f := &file{Buffer: bytes.NewBuffer(nil), name: "test.yaml"}
		assert.Error(t, packer.Pack(f, map[string]string{"foo": "bar"}))

		var data map[string]string
		assert.Error(t, packer.Unpack(f, &data))
	})
}

type file struct {
	*bytes.Buffer
	name string
}

func (f *file) Name() string { return f.name }
