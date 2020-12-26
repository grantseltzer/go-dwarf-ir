package gotir

import "fmt"

type gotir struct {
	Functions []*function_type
	Structs   []*struct_type
}

type go_type interface {
	String() string
}

type struct_type struct {
	Name   string
	Size   int64
	Fields []struct_field
}

func (s *struct_type) String() string {
	return s.Name
}

// Note: Can't embed struct_type because we can't allow to be recursive
// all types are indexed anyway, can lookup
type struct_field struct {
	Name     string
	TypeName string
	Offset   int64
}

func (f *struct_field) String() string {
	return fmt.Sprintf("%s %s %s\n", f.Name, f.Offset, f.TypeName)
}

type function_type struct {
	Name   string
	Params []go_type
}

type interface_type struct{}
