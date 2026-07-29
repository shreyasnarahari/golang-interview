package practice

type Speaker interface {
	Speak() string
}

type Dog struct {
	Name string
}

// Pointer receiver method
func (d *Dog) Speak() string {
	return "Woof! I am " + d.Name
}

// InterfaceCheck verifies if an interface contains a nil pointer.
func IsInterfaceDataNil(s Speaker) bool {
	return s == nil
}
