package file

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
)

type Output interface {
	io.Writer
	Name() string
}

type Encoder interface {
	Encode(any) error
}

type Input interface {
	io.Reader
	Name() string
}

type Decoder interface {
	Decode(any) error
}

type Packer interface {
	Register(func(io.Writer) Encoder, func(io.Reader) Decoder, ...string)
	Pack(Output, any) error
	Unpack(Input, any) error
}

func NewPacker() *packer {
	return &packer{
		encoders: make(map[string]func(io.Writer) Encoder),
		decoders: make(map[string]func(io.Reader) Decoder),
	}
}

type packer struct {
	guard    sync.RWMutex
	encoders map[string]func(io.Writer) Encoder
	decoders map[string]func(io.Reader) Decoder
}

func (r *packer) Register(
	encoder func(io.Writer) Encoder,
	decoder func(io.Reader) Decoder,
	extensions ...string,
) {
	r.guard.Lock()
	for _, ext := range extensions {
		r.encoders[ext] = encoder
		r.decoders[ext] = decoder
	}
	r.guard.Unlock()
}

func (r *packer) Pack(file Output, data any) error {
	ext := strings.ToLower(filepath.Ext(file.Name()))

	r.guard.RLock()
	encoder, has := r.encoders[ext]
	r.guard.RUnlock()

	if has {
		return encoder(file).Encode(data)
	}
	return fmt.Errorf("unsupported format: %s", ext)
}

func (r *packer) Unpack(file Input, ptr any) error {
	ext := strings.ToLower(filepath.Ext(file.Name()))

	r.guard.RLock()
	decoder, has := r.decoders[ext]
	r.guard.RUnlock()

	if has {
		return decoder(file).Decode(ptr)
	}
	return fmt.Errorf("unsupported format: %s", ext)
}
