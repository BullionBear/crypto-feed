package linkedlist

import "testing"

func TestNewIndexedLinkedList(t *testing.T) {
	ll := NewIndexedLinkedList[int]()
	if ll.size != 0 {
		t.Errorf("Expected initial size of 0, got %d", ll.size)
	}
	if len(ll.nodeIndex) != 0 {
		t.Errorf("Expected initial nodeIndex to be empty, got size %d", len(ll.nodeIndex))
	}
}

func TestPushBackAndTail(t *testing.T) {
	ll := NewIndexedLinkedList[int]()
	err := ll.PushBack(1, 100)
	if err != nil {
		t.Errorf("Error pushing back: %v", err)
	}

	if ll.size != 1 {
		t.Errorf("Expected size of 1 after one push, got %d", ll.size)
	}

	tail, err := ll.Tail()
	if err != nil {
		t.Errorf("Error getting tail: %v", err)
	}

	if tail != 100 {
		t.Errorf("Expected tail data of 100, got %d", tail)
	}
}

func TestPushFrontAndHead(t *testing.T) {
	ll := NewIndexedLinkedList[int]()
	err := ll.PushFront(2, 200)
	if err != nil {
		t.Errorf("Error pushing front: %v", err)
	}

	if ll.size != 1 {
		t.Errorf("Expected size of 1 after one push, got %d", ll.size)
	}

	head, err := ll.Head()
	if err != nil {
		t.Errorf("Error getting head: %v", err)
	}

	if head != 200 {
		t.Errorf("Expected head data of 200, got %d", head)
	}
}

func TestErrorOnDuplicateIndex(t *testing.T) {
	ll := NewIndexedLinkedList[int]()
	_ = ll.PushBack(1, 100)
	err := ll.PushBack(1, 101)
	if err != errIndexExist {
		t.Errorf("Expected error 'index is existed', got %v", err)
	}
}

func TestErrorOnEmptyListHeadTail(t *testing.T) {
	ll := NewIndexedLinkedList[int]()
	_, err := ll.Head()
	if err != errPtrIsNil {
		t.Errorf("Expected error 'pointer is nil' for empty list on Head, got %v", err)
	}

	_, err = ll.Tail()
	if err != errPtrIsNil {
		t.Errorf("Expected error 'pointer is nil' for empty list on Tail, got %v", err)
	}
}

func TestPopBack(t *testing.T) {
	ll := NewIndexedLinkedList[int]()
	_, err := ll.PopBack()
	if err != errPtrIsNil {
		t.Errorf("Expected error 'pointer is nil' for PopBack on empty list, got %v", err)
	}

	ll.PushBack(1, 100)
	ll.PushBack(2, 200)

	data, err := ll.PopBack()
	if err != nil {
		t.Errorf("Unexpected error on PopBack: %v", err)
	}
	if data != 200 {
		t.Errorf("Expected to pop back 200, got %d", data)
	}
	if ll.size != 1 {
		t.Errorf("Expected size 1 after popping back, got %d", ll.size)
	}
}

func TestPopFront(t *testing.T) {
	ll := NewIndexedLinkedList[int]()
	_, err := ll.PopFront()
	if err != errPtrIsNil {
		t.Errorf("Expected error 'pointer is nil' for PopFront on empty list, got %v", err)
	}

	ll.PushFront(1, 100)
	ll.PushFront(2, 200)

	data, err := ll.PopFront()
	if err != nil {
		t.Errorf("Unexpected error on PopFront: %v", err)
	}
	if data != 200 {
		t.Errorf("Expected to pop front 200, got %d", data)
	}
	if ll.size != 1 {
		t.Errorf("Expected size 1 after popping front, got %d", ll.size)
	}
}

func TestPopIndex(t *testing.T) {
	ll := NewIndexedLinkedList[int]()
	_, err := ll.PopIndex(1)
	if err != errIndexNotExist {
		t.Errorf("Expected error 'index is not existed' for non-existent index, got %v", err)
	}

	ll.PushBack(1, 100)
	ll.PushBack(2, 200)
	data, err := ll.PopIndex(2)
	if err != nil {
		t.Errorf("Unexpected error on PopIndex: %v", err)
	}
	if data != 200 {
		t.Errorf("Expected to pop index 200, got %d", data)
	}
	if _, exists := ll.nodeIndex[2]; exists {
		t.Errorf("Expected index 2 to be deleted from nodeIndex")
	}
	if ll.size != 1 {
		t.Errorf("Expected size 1 after popping index, got %d", ll.size)
	}
}

func TestGet(t *testing.T) {
	ll := NewIndexedLinkedList[int]()
	_, err := ll.Get(1)
	if err != errIndexNotExist {
		t.Errorf("Expected error 'index is not existed' for Get on non-existent index, got %v", err)
	}

	ll.PushBack(1, 100)
	data, err := ll.Get(1)
	if err != nil {
		t.Errorf("Unexpected error on Get: %v", err)
	}
	if data != 100 {
		t.Errorf("Expected to get data 100, got %d", data)
	}
}

func TestSize(t *testing.T) {
	ll := NewIndexedLinkedList[int]()
	if ll.Size() != 0 {
		t.Errorf("Expected size 0 for new list, got %d", ll.Size())
	}

	ll.PushBack(1, 100)
	if ll.Size() != 1 {
		t.Errorf("Expected size 1 after one push, got %d", ll.Size())
	}
}
