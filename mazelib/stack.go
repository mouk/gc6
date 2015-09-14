package mazelib

import "errors"

//Stack reprents a LiFO data type
type Stack struct {
	data []interface{}
}

var ErrEmptyStack = errors.New("Stack is empty")

//NewStack creates an return reference to a new stack with the specified capacity
func NewStack(capacity uint) *Stack {
	return &Stack{data: make([]interface{}, 0, capacity)}
}

//Len retrieves the length of the stack
func (s *Stack) Len() int {
	return len(s.data)
}

//Push new item to the top of the stack
func (s *Stack) Push(value interface{}) {
	s.data = append(s.data, value)
}

//Pop the last added item
func (s *Stack) Pop() (interface{}, error) {
	if s.Len() > 0 {
		ret := s.data[s.Len()-1]
		s.data = s.data[:s.Len()-1]
		return ret, nil
	}
	return nil, ErrEmptyStack
}
