package gotir

import "fmt"

type Gotir struct {
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

// Note: Can't embed structs/functions because we can't allow to be recursive
// all types are indexed anyway
type struct_field struct {
	Name     string
	TypeName string
	Offset   int64
}

func (f *struct_field) String() string {
	return fmt.Sprintf("%s %s %s", f.Name, f.Offset, f.TypeName)
}

type function_type struct {
	Name   string
	Params []function_param
}

func (f *function_type) String() string {
	return fmt.Sprintf("%s %v", f.Name, f.Params)
}

type function_param struct {
	Name string
}

type interface_type struct{}
