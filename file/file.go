package file

import (
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidFile   = errors.New("invalid file")
	ErrUnexpectedEOF = errors.New("unexpected end of file")
)

type File struct {
	BSS   uint64
	Stack uint64
	Image []byte
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

	file.Image = self.file[self.index:]

	return file, nil
}
