package binaryfmt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	Magic        = "FSN4"
	FormatV1     = uint16(1)
	DefaultMemSz = uint32(64 * 1024)
)

type Image struct {
	Version          uint16
	MemorySize       uint32
	CodeBase         uint32
	DataBase         uint32
	EntryPoint       uint32
	InputHandlerAddr uint32
	Code             []byte
	Data             []byte
}

type header struct {
	Magic            [4]byte
	Version          uint16
	Reversed         uint16
	MemorySize       uint32
	CodeBase         uint32
	DataBase         uint32
	EntryPoint       uint32
	InputHandlerAddr uint32
	CodeSize         uint32
	DataSize         uint32
}

func WriteFile(path string, img Image) error {
	var h header
	copy(h.Magic[:], []byte(Magic))
	if img.Version == 0 {
		img.Version = FormatV1
	}
	h.Version = img.Version
	if img.MemorySize == 0 {
		img.MemorySize = DefaultMemSz
	}
	h.MemorySize = img.MemorySize
	h.CodeBase = img.CodeBase
	h.DataBase = img.DataBase
	h.EntryPoint = img.EntryPoint
	h.InputHandlerAddr = img.InputHandlerAddr
	h.CodeSize = uint32(len(img.Code))
	h.DataSize = uint32(len(img.Data))

	buf := bytes.NewBuffer(nil)

	if err := binary.Write(buf, binary.LittleEndian, h); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}
	if _, err := buf.Write(img.Code); err != nil {
		return fmt.Errorf("append code: %w", err)
	}
	if _, err := buf.Write(img.Data); err != nil {
		return fmt.Errorf("append data: %w", err)
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func ReadFile(path string) (Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Image{}, err
	}
	return Parse(data)
}

func Parse(data []byte) (Image, error) {
	r := bytes.NewReader(data)
	var h header
	if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
		return Image{}, fmt.Errorf("decode header: %w", err)
	}
	if string(h.Magic[:]) != Magic {
		return Image{}, errors.New("invalid magic")
	}
	if h.Version != FormatV1 {
		return Image{}, fmt.Errorf("unsupported version: %d", h.Version)
	}
	code := make([]byte, h.CodeSize)
	if _, err := io.ReadFull(r, code); err != nil {
		return Image{}, fmt.Errorf("decode code: %w", err)
	}
	blob := make([]byte, h.DataSize)
	if _, err := io.ReadFull(r, blob); err != nil {
		return Image{}, fmt.Errorf("decode data: %w", err)
	}
	return Image{
		Version:          h.Version,
		MemorySize:       h.MemorySize,
		CodeBase:         h.CodeBase,
		DataBase:         h.DataBase,
		EntryPoint:       h.EntryPoint,
		InputHandlerAddr: h.InputHandlerAddr,
		Code:             code,
		Data:             blob,
	}, nil
}
