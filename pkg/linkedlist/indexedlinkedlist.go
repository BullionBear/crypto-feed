package linkedlist

import (
	"errors"
)

var (
	errPtrIsNil      = errors.New("pointer is nil")
	errIndexExist    = errors.New("index is existed")
	errIndexNotExist = errors.New("index is not existed")
)

type IndexedNode[T any] struct {
	prev  *IndexedNode[T]
	next  *IndexedNode[T]
	index int64
	data  T
}

type IndexLinkedList[T any] struct {
	head      *IndexedNode[T]
	tail      *IndexedNode[T]
	size      int64
	nodeIndex map[int64]*IndexedNode[T]
}

func NewIndexedLinkedList[T any]() *IndexLinkedList[T] {
	return &IndexLinkedList[T]{
		size:      0,
		nodeIndex: make(map[int64]*IndexedNode[T]),
	}
}

func (ls *IndexLinkedList[T]) Head() (T, error) {
	if ls.size == 0 {
		var zero T
		return zero, errPtrIsNil
	}
	return ls.head.data, nil
}

func (ls *IndexLinkedList[T]) HeadKey(step int) (int64, error) {
	ptr := ls.head
	if ptr == nil {
		return 0, errPtrIsNil
	}
	for i := 0; i < step; i++ {
		if ptr != nil && ptr.next != nil {
			ptr = ptr.next
		} else {
			return 0, errPtrIsNil
		}
	}
	return ptr.index, nil
}

func (ls *IndexLinkedList[T]) Tail() (T, error) {
	if ls.size == 0 {
		var zero T
		return zero, errPtrIsNil
	}
	return ls.tail.data, nil
}

func (ls *IndexLinkedList[T]) TailKey(step int) (int64, error) {
	ptr := ls.tail
	if ptr == nil {
		return 0, errPtrIsNil
	}
	for i := 0; i < step; i++ {
		if ptr != nil && ptr.prev != nil {
			ptr = ptr.prev
		} else {
			return 0, errPtrIsNil
		}
	}
	return ptr.index, nil
}

func (ls *IndexLinkedList[T]) PushBack(index int64, data T) error {
	if ls.size == 0 {
		return ls.insertFirstNode(index, data)
	}
	if _, exists := ls.nodeIndex[index]; exists {
		return errIndexExist
	}
	var newNode = IndexedNode[T]{
		data:  data,
		index: index,
		prev:  ls.tail,
	}
	ls.tail.next = &newNode
	ls.tail = &newNode
	ls.nodeIndex[index] = &newNode
	ls.size++
	return nil
}

func (ls *IndexLinkedList[T]) PushFront(index int64, data T) error {
	if ls.size == 0 {
		return ls.insertFirstNode(index, data)
	}
	if _, exists := ls.nodeIndex[index]; exists {
		return errIndexExist
	}

	newNode := &IndexedNode[T]{
		data:  data,
		index: index,
		next:  ls.head,
	}

	ls.head.prev = newNode
	ls.head = newNode
	ls.nodeIndex[index] = newNode
	ls.size++
	return nil
}

func (ls *IndexLinkedList[T]) PopBack() (T, error) {
	if ls.size == 0 {
		var zero T
		return zero, errPtrIsNil
	}

	data := ls.tail.data
	delete(ls.nodeIndex, ls.tail.index)

	if ls.size == 1 {
		ls.head = nil
		ls.tail = nil
	} else {
		ls.tail = ls.tail.prev
		ls.tail.next = nil
	}

	ls.size--

	return data, nil
}

func (ls *IndexLinkedList[T]) PopFront() (T, error) {
	if ls.size == 0 {
		var zero T
		return zero, errPtrIsNil
	}

	data := ls.head.data
	delete(ls.nodeIndex, ls.head.index)

	if ls.size == 1 {
		ls.head = nil
		ls.tail = nil
	} else {
		ls.head = ls.head.next
		ls.head.prev = nil
	}
	ls.size--

	return data, nil
}

func (ls *IndexLinkedList[T]) PopIndex(index int64) (T, error) {
	// Check if the node exists in the map
	node, exists := ls.nodeIndex[index]
	if !exists {
		var zero T
		return zero, errIndexNotExist
	}

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		ls.head = node.next
	}
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		ls.tail = node.prev
	}

	delete(ls.nodeIndex, index)
	ls.size--
	return node.data, nil
}

func (ls *IndexLinkedList[T]) Get(index int64) (T, error) {
	if node, exists := ls.nodeIndex[index]; exists {
		return node.data, nil
	}
	var zero T
	return zero, errIndexNotExist
}

func (ls *IndexLinkedList[T]) Size() int64 {
	return ls.size
}

func (ls *IndexLinkedList[T]) Next(index int64) (int64, error) {
	node, exists := ls.nodeIndex[index]
	if !exists {
		return 0, errIndexNotExist
	}
	if node.next != nil {
		return node.next.index, nil
	}
	return 0, errPtrIsNil
}

func (ls *IndexLinkedList[T]) Prev(index int64) (int64, error) {
	node, exists := ls.nodeIndex[index]
	if !exists {
		return 0, errIndexNotExist
	}
	if node.prev != nil {
		return node.prev.index, nil
	}
	return 0, errPtrIsNil
}

func (ls *IndexLinkedList[T]) insertFirstNode(index int64, data T) error {
	var firstNode = IndexedNode[T]{
		data:  data,
		index: index,
	}
	ls.head = &firstNode
	ls.tail = &firstNode
	ls.nodeIndex[index] = &firstNode
	ls.size++
	return nil
}
