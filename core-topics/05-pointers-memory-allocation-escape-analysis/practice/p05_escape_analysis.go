package practice

type UserData struct {
	ID   int
	Name string
}

// StackAllocated returns value by copy (stays on stack).
func StackAllocated(id int, name string) UserData {
	return UserData{ID: id, Name: name}
}

// HeapAllocated returns pointer to local struct (escapes to heap!).
func HeapAllocated(id int, name string) *UserData {
	u := UserData{ID: id, Name: name}
	return &u // Escapes to heap because pointer is returned
}
