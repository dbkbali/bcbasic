package core

import "fmt"

type Instruction byte

const (
	InstrPushInt  Instruction = 0x0a
	InstrAdd      Instruction = 0x0b
	InstrPushByte Instruction = 0x0c
	InstrPack     Instruction = 0x0d
	InstrSub      Instruction = 0x0e
)

type Stack struct {
	data []any
	sPtr int
}

func NewStack(size int) *Stack {
	return &Stack{
		data: make([]any, size),
		sPtr: 0,
	}
}

func (s *Stack) Push(a any) {
	s.data[s.sPtr] = a
	s.sPtr++
}

func (s *Stack) Pop() any {
	value := s.data[0]
	s.data = append(s.data[:0], s.data[1:]...)
	s.sPtr--

	return value
}

type VM struct {
	data    []byte
	instPtr int
	stack   *Stack
}

func NewVM(data []byte) *VM {
	return &VM{
		data:    data,
		instPtr: 0,
		stack:   NewStack(128),
	}
}

func (vm *VM) Run() error {
	for {
		instr := Instruction(vm.data[vm.instPtr])

		fmt.Println(instr)

		if err := vm.Exec(instr); err != nil {
			return err
		}

		vm.instPtr++

		if vm.instPtr >= len(vm.data) {
			break
		}
	}
	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case InstrPushInt:
		vm.stack.Push(int(vm.data[vm.instPtr-1]))

	case InstrPushByte:
		vm.stack.Push(byte(vm.data[vm.instPtr-1]))

	case InstrPack:
		n := vm.stack.Pop().(int)
		b := make([]byte, n)
		for i := 0; i < n; i++ {
			b[i] = vm.stack.Pop().(byte)
		}

		vm.stack.Push(b)

	case InstrSub:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a - b
		vm.stack.Push(c)

	case InstrAdd:
		a := vm.stack.Pop().(int)
		b := vm.stack.Pop().(int)
		c := a + b
		vm.stack.Push(c)
	}
	return nil
}
