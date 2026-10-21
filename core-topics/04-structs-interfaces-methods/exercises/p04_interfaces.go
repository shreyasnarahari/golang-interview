package exercises

type Speaker interface {
	Speak() string
}

type Dog struct {
	Name string
}

// Pointer receiver method
func (d *Dog) Speak() string {
	panic("TODO: implement Speak")
}

// InterfaceCheck verifies if an interface contains a nil pointer.
func IsInterfaceDataNil(s Speaker) bool {
	panic("TODO: implement IsInterfaceDataNil")
}
