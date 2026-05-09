package file

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
)

var (
	ErrInvalidFile   = errors.New("invalid file")
	ErrUnexpectedEOF = errors.New("unexpected end of file")
)

type Label struct {
	Offset int
	Size   int
}

type File struct {
	BSS   uint64
	Stack uint64
	Image []byte

	MemorySize uint64
	MemoryMap  map[string]Label
}

type Parser struct {
	file  []byte
	index int
}

func NewParser(file []byte) Parser {
	return Parser{
		file:  file,
		index: 0,
	}
}

func (self *Parser) ReadBytes(size int) ([]byte, bool) {
	if (self.index + size) >= len(self.file) {
		return nil, false
	}

	bytes := self.file[self.index:(self.index + size)]
	self.index += size
	return bytes, true
}

func (self *Parser) ReadUint64() (uint64, bool) {
	bytes, ok := self.ReadBytes(8)
	if !ok {
		return 0, ok
	}

	return binary.BigEndian.Uint64(bytes), true
}

func (self *Parser) Parse() (File, error) {
	file := File{}

	voran, ok := self.ReadBytes(5)
	if !ok {
		return File{}, ErrUnexpectedEOF
	}

	if string(voran) != "VORAN" {
		return File{}, ErrInvalidFile
	}

	bss, ok := self.ReadUint64()
	if !ok {
		return File{}, ErrUnexpectedEOF
	}
	file.BSS = bss

	stack, ok := self.ReadUint64()
	if !ok {
		return File{}, ErrUnexpectedEOF
	}
	file.Stack = stack

	memorySize, ok := self.ReadUint64()
	if !ok {
		return File{}, ErrUnexpectedEOF
	}
	file.MemorySize = memorySize

	if memorySize > 0 {
		memoryMapBytes, ok := self.ReadBytes(int(memorySize))
		if !ok {
			return File{}, ErrUnexpectedEOF
		}

		var decoderBuffer bytes.Buffer
		_, err := decoderBuffer.Write(memoryMapBytes)
		if err != nil {
			return File{}, err
		}

		decoder := gob.NewDecoder(&decoderBuffer)

		var memoryMap map[string]Label

		err = decoder.Decode(&memoryMap)
		if err != nil {
			return File{}, err
		}

		file.MemoryMap = memoryMap
	}

	file.Image = self.file[self.index:]

	return file, nil
}
