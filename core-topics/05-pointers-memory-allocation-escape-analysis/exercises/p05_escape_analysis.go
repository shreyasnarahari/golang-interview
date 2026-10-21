package exercises

type UserData struct {
	ID   int
	Name string
}

// StackAllocated returns value by copy (stays on stack).
func StackAllocated(id int, name string) UserData {
	panic("TODO: implement StackAllocated")
}

// HeapAllocated returns pointer to local struct (escapes to heap!).
func HeapAllocated(id int, name string) *UserData {
	panic("TODO: implement HeapAllocated")
}
