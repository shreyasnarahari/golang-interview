package solutions

type UserData struct {
	ID   int
	Name string
}

// heapSink forces any pointer assigned to it to escape to the heap.
var heapSink *UserData

// StackAllocated returns value by copy (stays on stack).
func StackAllocated(id int, name string) UserData {
	return UserData{ID: id, Name: name}
}

// HeapAllocated returns pointer to local struct (escapes to heap!).
func HeapAllocated(id int, name string) *UserData {
	u := UserData{ID: id, Name: name}
	heapSink = &u // Force escape: pointer assigned to package-level sink
	return &u     // Escapes to heap because pointer is returned
}
