package core

import (
	"fmt"
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

func TestVMCalculateStore(t *testing.T) {
	contractState := NewState()
	data := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	dataOther := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4d, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}

	data = append(data, dataOther...)

	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	fmt.Printf("%+v\n", vm.contractState)

	valueBytes, err := vm.contractState.Get([]byte("FOO"))
	assert.Nil(t, err)
	value := deserializeInt64(valueBytes)
	assert.Equal(t, int64(5), value)
}

func TestVM2(t *testing.T) {

	data := []byte{0x02, 0x0a, 0x03, 0x0a, 0x0b, 0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}
	pushFoo := []byte{0x4f, 0x0c, 0x4f, 0x0c, 0x46, 0x0c, 0x03, 0x0a, 0x0d, 0xae}
	data = append(data, pushFoo...)

	contractState := NewState()
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	value := vm.stack.Pop().([]byte)
	valueSerialized := deserializeInt64(value)

	assert.Equal(t, int64(5), valueSerialized)
	fmt.Printf("%+v\n", vm.stack.data)

}

func TestVMMul(t *testing.T) {
	contractState := NewState()
	data := []byte{0x02, 0x0a, 0x02, 0x0a, 0x0ea}
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop().(int)

	assert.Equal(t, 4, result)
}

func TestVMDiv(t *testing.T) {
	contractState := NewState()
	data := []byte{0x08, 0x0a, 0x02, 0x0a, 0x0fd}
	vm := NewVM(data, contractState)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop().(int)

	assert.Equal(t, 4, result)
}
