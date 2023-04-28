package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	s := NewStack(128)

	s.Push(1)
	s.Push(2)

	assert.Equal(t, 2, s.sPtr)
	assert.Equal(t, 1, s.data[0])
	assert.Equal(t, 2, s.data[1])
	value := s.Pop()

	assert.Equal(t, 1, value)

	value = s.Pop()
	assert.Equal(t, 2, value)

	// assert.Equal(t, 3, s.sPtr)
	// assert.Equal(t, 3, s.Pop())
	// assert.Equal(t, 2, s.Pop())
	// assert.Equal(t, 1, s.Pop())
	// assert.Equal(t, -1, s.sPtr)
}

func TestVM(t *testing.T) {
	// 1 + 2 = 3
	// 1
	// push stack
	// 2
	// push stack
	// add
	contractState := NewState()
	data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d}
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop().([]byte)

	assert.Equal(t, "FOO", string(result))
}

func TestVMSub(t *testing.T) {
	// 1 + 2 = 3
	// 1
	// push stack
	// 2
	// push stack
	// add
	contractState := NewState()
	data := []byte{0x03, 0x0a, 0x02, 0x0a, 0x0e}
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop().(int)

	assert.Equal(t, 1, result)
}
func TestVMStore(t *testing.T) {
	contractState := NewState()
	data := []byte{0x03, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x0d, 0x05, 0x0a, 0xf}
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	v, err := contractState.Get([]byte("FOO"))
	value := deserializeInt64(v)
	assert.Nil(t, err)
	assert.Equal(t, value, int64(5))
}
