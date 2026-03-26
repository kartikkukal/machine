package processor

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var (
	ErrFellOff           = errors.New("fell off memory")
	ErrUnkownRegister    = errors.New("unknown register")
	ErrMemoryOutOfBounds = errors.New("memory out of bounds")
)

type Instruction struct {
	code   uint64
	kind   uint64
	first  uint64
	second uint64
}

type Processor struct {
	registers []uint64

	overflow bool
	carry    bool
	negative bool
	zero     bool
	positive bool

	memory []byte
}

func New(memory []byte) Processor {
	return Processor{
		registers: make([]uint64, 16),
		memory:    memory,
	}
}

func twosComplement(u uint64, n int) int64 {
	mask := uint64(1 << (n - 1))

	var i int64

	i -= int64(u & mask)
	i += int64(u & (^mask))

	return int64(i)
}

func parseInstruction(instruction uint64) Instruction {
	t := instruction >> 46
	code := instruction >> 40

	var first uint64
	var second uint64

	switch t {
	case 0:
		first = (instruction & 0x000000FC00000000) >> 34
		second = (instruction & 0x00000003FFFFFFFF)
	case 1:
		first = (instruction & 0x000000FC00000000) >> 34
		second = (instruction & 0x00000003FFFFFFFF)
	case 2:
		first = (instruction & 0x000000FC00000000) >> 34
		second = (instruction & 0x00000003F0000000) >> 28
	case 3:
		first = (instruction & 0x000000FFFFFFFFFF)
	}

	return Instruction{
		code:   code & 0x3F,
		kind:   t,
		first:  first,
		second: second,
	}
}

func (self *Processor) getRegister(index uint64) (uint64, error) {
	if index >= 64 {
		return 0, ErrUnkownRegister
	}

	index = index & 0x0F

	return self.registers[index], nil
}

func (self *Processor) getMemory(index uint64, length uint64) (uint64, error) {
	size := uint64(1)

	if length > 0 {
		size = 2 << (length - 1)
	}

	if index+size > uint64(len(self.memory)) {
		return 0, ErrMemoryOutOfBounds
	}

	data := make([]byte, 8-(size))
	data = append(data, self.memory[index:index+(size)]...)

	return binary.BigEndian.Uint64(data), nil
}

func (self *Processor) setRegister(index uint64, value uint64) error {
	if index >= 64 {
		return ErrUnkownRegister
	}

	index = index & 0x0F
	self.registers[index] = value

	return nil
}

func (self *Processor) setMemory(index uint64, length uint64, value uint64) error {
	size := uint64(1)

	if length > 0 {
		size = 2 << (length - 1)
	}

	if index+size > uint64(len(self.memory)) {
		return ErrMemoryOutOfBounds
	}

	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, value)

	for i := uint64(0); i < (size); i++ {
		self.memory[index+i] = data[8-(size)+i]
	}

	return nil
}

func (self *Processor) memorySize() uint64 {
	return uint64(len(self.memory))
}

func (self *Processor) getInstruction() (Instruction, error) {

	index := self.registers[15]

	if index+6 > self.memorySize() {
		return Instruction{}, ErrFellOff
	}

	instruction := make([]byte, 2)
	instruction = append(instruction, self.memory[index:index+6]...)

	return parseInstruction(binary.BigEndian.Uint64(instruction)), nil
}

func (self *Processor) Execute() error {

	wrapper := func(err error) error {
		return fmt.Errorf("error while executing: %w at memory index %v", err, self.registers[15])
	}

	for {
		i, err := self.getInstruction()

		if err != nil {
			return wrapper(err)
		}

		if i.code <= 2 {
			err := self.moveOperations(i)
			if err != nil {
				return wrapper(err)
			}
		}

		if i.code >= 3 && i.code <= 6 {
			err := self.bitwiseOperations(i)
			if err != nil {
				return wrapper(err)
			}
		}

		if i.code >= 9 && i.code <= 13 {
			err := self.arithmeticOperations(i)
			if err != nil {
				return wrapper(err)
			}
		}

		if i.code == 14 {
			err := self.comparisonOperations(i)
			if err != nil {
				return wrapper(err)
			}
		}

		if i.code >= 15 && i.code <= 21 {
			skip, err := self.jumpOperations(i)
			if err != nil {
				return wrapper(err)
			}

			if skip {
				continue
			}
		}

		if i.code == 22 {
			ret, err := self.handleInterrupt(i)
			if err != nil {
				return wrapper(err)
			}

			if ret {
				return nil
			}
		}

		self.registers[15] += 6
	}
}
