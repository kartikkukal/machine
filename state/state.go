package state

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type Instruction struct {
	code   uint64
	kind   uint64
	first  uint64
	second uint64
}

type State struct {
	registers []uint64

	overflow bool
	carry    bool
	negative bool
	zero     bool
	positive bool

	memory []byte
}

var (
	ErrFellOff            = errors.New("fell off memory")
	ErrExecution          = errors.New("error while executing instruction")
	ErrUnknownInstruction = errors.New("unknown instruction")
)

func Pow(x int, y uint64) uint64 {
	if y == 0 {
		return 1
	}

	base := x

	for i := uint64(1); i < y; i++ {
		base *= x
	}

	return uint64(base)
}

func New(memory []byte) State {
	return State{
		registers: make([]uint64, 16),
		memory:    memory,
	}
}

func complement(u uint64) int64 {
	i := int32(u)
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

func (self *State) getRegister(index uint64) (uint64, bool) {
	if index >= 64 {
		return 0, false
	}

	index = index & 0x0F

	return self.registers[index], true
}

func (self *State) getMemory(index uint64, length uint64) (uint64, bool) {
	size := Pow(2, length)

	if index >= uint64(len(self.memory)) {
		return 0, false
	}

	if index+size > uint64(len(self.memory)) {
		return 0, false
	}

	data := make([]byte, 8-(size))
	data = append(data, self.memory[index:index+(size)]...)

	return binary.BigEndian.Uint64(data), true
}

func (self *State) setRegister(index uint64, value uint64) bool {
	if index >= 64 {
		return false
	}

	index = index & 0x0F

	self.registers[index] = value
	return true
}

func (self *State) setMemory(index uint64, length uint64, value uint64) bool {
	size := Pow(2, length)

	if index >= uint64(len(self.memory)) {
		return false
	}

	if index+size > uint64(len(self.memory)) {
		return false
	}

	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, value)

	for i := uint64(0); i < (size); i++ {
		self.memory[index+i] = data[8-(size)+i]
	}

	return true
}

func (self *State) MoveOperations(i Instruction) bool {

	switch i.code {
	case 0:
		value, ok := self.getRegister(i.first)
		if !ok {
			return false
		}

		switch i.kind {
		case 0:
			ok = self.setMemory(i.second, i.first>>4, value)
			if !ok {
				return false
			}
		case 2:
			address, ok := self.getRegister(i.second)
			if !ok {
				return false
			}

			ok = self.setMemory(address, i.first>>4, value)
			if !ok {
				return false
			}
		}
	case 1:

		var value uint64
		var ok bool

		switch i.kind {
		case 0:
			value, ok = self.getMemory(i.second, i.first>>4)
			if !ok {
				return false
			}

		case 1:
			value = i.second

		case 2:
			address, ok := self.getRegister(i.second)
			if !ok {
				return false
			}

			value, ok = self.getMemory(address, i.first>>4)
			if !ok {
				return false
			}
		}

		ok = self.setRegister(i.first, value)
		if !ok {
			return false
		}
	case 2:

		value, ok := self.getRegister(i.first)
		if !ok {
			return false
		}

		ok = self.setRegister(i.second, value)
		if !ok {
			return false
		}
	}

	return true
}

func (self *State) BitwiseOperations(i Instruction) bool {

	value_1, ok := self.getRegister(i.first)
	if !ok {
		return false
	}

	value_2, ok := self.getRegister(i.second)
	if !ok {
		return false
	}

	var result uint64

	switch i.code {
	case 3:
		result = value_1
	case 4:
		result = value_1 & value_2
	case 5:
		result = value_1 | value_2
	case 6:
		result = value_1 ^ value_2
	}

	return self.setRegister(i.second, result)
}

func (self *State) ArithmeticOperations(i Instruction) bool {
	value_1, ok := self.getRegister(i.first)
	if !ok {
		return false
	}

	var value_2 uint64

	switch i.kind {
	case 1:
		value_2 = i.second
	case 2:
		value_2, ok = self.getRegister(i.second)
		if !ok {
			return false
		}
	}

	switch i.code {
	case 9:
		return self.setRegister(i.first, value_1+value_2)
	case 10:
		return self.setRegister(i.first, value_1-value_2)
	case 11:
		return self.setRegister(i.first, value_1*value_2)
	case 12:
		return self.setRegister(i.first, value_1/value_2)
	case 13:
		return self.setRegister(i.first, value_1%value_2)
	}

	return false
}

func (self *State) ComparisonOperation(i Instruction) bool {
	value_1, ok := self.getRegister(i.first)
	if !ok {
		return false
	}

	var value_2 uint64

	switch i.kind {
	case 1:
		value_2 = i.second
	case 2:
		value_2, ok = self.getRegister(i.second)
		if !ok {
			return false
		}
	}

	self.zero = false
	self.negative = false
	self.positive = false

	if value_1 == value_2 {
		self.zero = true
	}

	if value_1 < value_2 {
		self.negative = true
	}

	if value_1 > value_2 {
		self.positive = true
	}

	return true
}

func (self *State) JumpOperations(i Instruction) (bool, bool) {

	var value uint64
	var ok bool

	switch i.kind {
	case 1:
		value, ok = self.getRegister(i.first)
		if !ok {
			return false, false
		}

		self.registers[15] = i.second

	case 3:
		value = uint64(complement(i.first))
	}

	switch i.code {
	case 15:
		if !self.zero {
			self.registers[15] += value
			return true, true
		}

	case 16:
		if self.negative {
			self.registers[15] += value
			return true, true
		}

	case 17:
		if self.zero || self.negative {
			self.registers[15] += value
			return true, true
		}

	case 18:
		if self.positive {
			self.registers[15] += value
			return true, true
		}

	case 19:
		if self.zero || self.positive {
			self.registers[15] += value
			return true, true
		}

	case 20:
		if self.zero {
			self.registers[15] += value
			return true, true
		}

	case 21:
		self.registers[15] += value
		return true, true
	}

	return false, true
}

func (self *State) InterruptPrint(address uint64) {

	output := make([]byte, 0)

	for {
		if address >= uint64(len(self.memory)) {
			break
		}

		value := self.memory[address]

		if value == 0 {
			break
		}

		output = append(output, value)

		address++
	}

	fmt.Println(string(output))
}

func (self *State) HandleInterrupts(i Instruction) (bool, bool) {
	if i.code != 22 {
		return false, false
	}

	switch i.first {
	case 0:
		return true, true

	case 1:
		address, ok := self.getRegister(0)
		if !ok {
			return false, false
		}

		self.InterruptPrint(address)

	case 2:
		length, ok := self.getRegister(0)
		if !ok {
			return false, false
		}

		ok = self.setRegister(0, uint64(len(self.memory)))
		if !ok {
			return false, false
		}

		new := make([]byte, length)
		self.memory = append(self.memory, new...)

	case 3:
		fmt.Println(self.registers)
	case 4:
		for i, b := range self.memory {
			fmt.Printf("%08b ", b)

			if (i+1)%8 == 0 {
				fmt.Println()
			}
		}
	}

	return false, true
}

func (self *State) Execute() error {

	for {

		if (self.registers[15] + 6) > uint64(len(self.memory)) {
			return ErrFellOff
		}

		slice := self.memory[self.registers[15] : self.registers[15]+6]
		slice = append([]byte{0, 0}, slice...)
		value := binary.BigEndian.Uint64(slice)

		i := parseInstruction(value)
		//fmt.Println(self.registers)
		//fmt.Println(i)

		/*for i, b := range self.memory {
			fmt.Printf("%08b ", b)

			if (i+1)%8 == 0 {
				fmt.Println()
			}
		}*/

		//fmt.Println("\n")

		if i.code <= 2 {
			ok := self.MoveOperations(i)
			if !ok {
				panic(0)
			}

		} else if i.code <= 6 {
			ok := self.BitwiseOperations(i)
			if !ok {
				panic(0)
			}

		} else if i.code >= 9 && i.code <= 13 {
			ok := self.ArithmeticOperations(i)
			if !ok {
				panic(0)
			}

		} else if i.code == 14 {
			ok := self.ComparisonOperation(i)
			if !ok {
				panic(0)
			}

		} else if i.code >= 15 && i.code <= 21 {
			skip, ok := self.JumpOperations(i)
			if !ok {
				panic(0)
			}

			if skip {
				continue
			}

		} else if i.code == 22 {
			ret, ok := self.HandleInterrupts(i)
			if !ok {
				panic(0)
			}

			if ret {
				return nil
			}

		} else {

			return ErrUnknownInstruction
		}

		self.registers[15] += 6
	}
}
