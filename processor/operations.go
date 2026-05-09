package processor

import (
	"encoding/binary"
	"fmt"
)

func (self *Processor) moveOperations(i Instruction) error {

	wrapper := func(err error) error {
		return fmt.Errorf("while performing move operations: %w", err)
	}

	switch i.code {
	case 0:
		value, err := self.getRegister(i.first)
		if err != nil {
			return wrapper(err)
		}

		switch i.kind {
		case 0:
			err = self.setMemory(i.second, i.first>>4, value)
			if err != nil {
				return wrapper(err)
			}

		case 2:
			address, err := self.getRegister(i.second)
			if err != nil {
				return wrapper(err)
			}

			err = self.setMemory(address+i.offset, i.first>>4, value)
			if err != nil {
				return wrapper(err)
			}
		}

	case 1:
		var value uint64
		var err error

		switch i.kind {
		case 0:
			value, err = self.getMemory(i.second, i.first>>4)
			if err != nil {
				return wrapper(err)
			}

		case 1:
			value = i.second

		case 2:
			address, err := self.getRegister(i.second)
			if err != nil {
				return wrapper(err)
			}

			value, err = self.getMemory(address+i.offset, i.first>>4)
			if err != nil {
				return wrapper(err)
			}
		}

		err = self.setRegister(i.first, value)
		if err != nil {
			return wrapper(err)
		}

	case 2:
		value, err := self.getRegister(i.first)
		if err != nil {
			return wrapper(err)
		}

		err = self.setRegister(i.second, value)
		if err != nil {
			return wrapper(err)
		}
	}

	return nil
}

func (self *Processor) bitwiseOperations(i Instruction) error {

	wrapper := func(err error) error {
		return fmt.Errorf("while performing bitwise operations: %w", err)
	}

	value_1, err := self.getRegister(i.first)
	if err != nil {
		return wrapper(err)
	}

	value_2, err := self.getRegister(i.second)
	if err != nil {
		return wrapper(err)
	}

	value_2 += i.offset

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

	return self.setRegister(i.first, result)
}

func (self *Processor) arithmeticOperations(i Instruction) error {

	wrapper := func(err error) error {
		return fmt.Errorf("while performing arithmetic operations: %w", err)
	}

	value_1, err := self.getRegister(i.first)
	if err != nil {
		return wrapper(err)
	}

	var value_2 uint64

	switch i.kind {
	case 1:
		value_2 = i.second

	case 2:
		value_2, err = self.getRegister(i.second)
		if err != nil {
			return wrapper(err)
		}
		value_2 += i.offset
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

	return nil
}

func (self *Processor) comparisonOperations(i Instruction) error {

	wrapper := func(err error) error {
		return fmt.Errorf("while performing comparison operations: %w", err)
	}

	value_1, err := self.getRegister(i.first)
	if err != nil {
		return wrapper(err)
	}

	var value_2 uint64

	switch i.kind {
	case 1:
		value_2 = i.second

	case 2:
		value_2, err = self.getRegister(i.second)
		if err != nil {
			return wrapper(err)
		}
		value_2 += i.offset
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

	return nil
}

func (self *Processor) jumpOperations(i Instruction) (bool, error) {

	wrapper := func(err error) error {
		return fmt.Errorf("while performing jump operations: %w", err)
	}

	var value uint64
	var err error

	switch i.kind {
	case 1:
		value, err = self.getRegister(i.first)
		if err != nil {
			return false, wrapper(err)
		}

		self.registers[15] = i.second

	case 3:
		value = uint64(twosComplement(i.first, 40))
	}

	switch i.code {
	case 34:
		if !self.zero {
			self.registers[15] += value
			return true, nil
		}

	case 35:
		if self.negative {
			self.registers[15] += value
			return true, nil
		}

	case 36:
		if self.zero || self.negative {
			self.registers[15] += value
			return true, nil
		}

	case 37:
		if self.positive {
			self.registers[15] += value
			return true, nil
		}

	case 38:
		if self.zero || self.positive {
			self.registers[15] += value
			return true, nil
		}

	case 39:
		if self.zero {
			self.registers[15] += value
			return true, nil
		}

	case 40:
		self.registers[15] += value
		return true, nil
	}

	return false, nil
}

func (self *Processor) interruptPrint(address uint64) {

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

func (self *Processor) handleInterrupt(i Instruction) (bool, error) {

	wrapper := func(err error) error {
		return fmt.Errorf("while handling interrupt: %w", err)
	}

	switch i.first {
	case 0:
		return true, nil

	case 1:
		address, err := self.getRegister(0)
		if err != nil {
			return false, wrapper(err)
		}

		self.interruptPrint(address)

	case 2:
		length, err := self.getRegister(0)
		if err != nil {
			return false, wrapper(err)
		}

		err = self.setRegister(0, uint64(len(self.memory)))
		if err != nil {
			return false, wrapper(err)
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
		fmt.Println()
	case 5:
		for name, label := range self.memoryMap {
			start := label.Offset
			end := label.Offset + label.Size

			data := self.memory[start:end]

			fmt.Printf("%v: ", name)

			for _, b := range data {
				fmt.Printf("%08b ", b)
			}

			size := len(data)

			if size == 1 || size == 2 ||
				size == 4 || size == 8 {
				padded := make([]byte, 8-(size))
				padded = append(padded, data...)

				number := binary.BigEndian.Uint64(padded)

				fmt.Printf("/ Decimal: %v", number)
			}

			fmt.Println()
		}
	}

	return false, nil
}

func (self *Processor) setOperations(i Instruction) error {
	if i.kind != 1 {
		return nil
	}

	switch i.code {
	case 42:
		if self.zero {
			self.setRegister(i.first, i.second)
		}
	case 43:
		if !self.zero {
			self.setRegister(i.first, i.second)
		}
	case 44:
		if self.negative {
			self.setRegister(i.first, i.second)
		}
	case 45:
		if self.positive {
			self.setRegister(i.first, i.second)
		}
	case 46:
		if self.negative || self.zero {
			self.setRegister(i.first, i.second)
		}
	case 47:
		if self.positive || self.zero {
			self.setRegister(i.first, i.second)
		}
	}

	return nil
}
