package gotir

import (
	"debug/dwarf"
	"debug/elf"
	"fmt"
	"io"
)

type go_type interface {
	String() string
}

type struct_type struct {
	name   string
	size   int64
	fields []struct_field
}

type struct_field struct {
	name     string
	typeName string
	offset   int64
}

func (f *struct_field) String() string {
	return fmt.Sprintf("%s %s %s\n", f.name, f.offset, f.typeName)
}

func (s *struct_type) String() string {
	return s.name
}

type interface_type struct{}

type go_function interface {
	String() string
}

func ParseFromPath(path string) error {
	elfFile, err := elf.Open(path)
	if err != nil {
		return err
	}

	data, err := elfFile.DWARF()
	if err != nil {
		return err
	}

	structs, err := parseStructs(data)
	if err != nil {
		return err
	}

	fmt.Println(structs)

	return nil
}

func parseStructs(data *dwarf.Data) ([]*struct_type, error) {

	lineReader := data.Reader()
	typeReader := data.Reader()

	structs := []*struct_type{}
	var currentlyReadingStruct *struct_type = nil

	for {
		entry, err := lineReader.Next()
		if err == io.EOF || entry == nil {
			break
		}
		if err != nil {
			return nil, err
		}
		if entry == nil {
			structs = append(structs, currentlyReadingStruct)
			currentlyReadingStruct = nil
		}

		// Found a struct
		if entry.Tag == dwarf.TagStructType {

			currentlyReadingStruct = &struct_type{}

			for _, field := range entry.Field {

				if field.Attr == dwarf.AttrName {
					currentlyReadingStruct.name = field.Val.(string)
				}

				if field.Attr == dwarf.AttrByteSize {
					currentlyReadingStruct.size = field.Val.(int64)
				}
			}

			currentlyReadingStruct.fields = []struct_field{}
		}

		// If currently reading the fields of a struct
		if currentlyReadingStruct != nil {

			// Found a member of the struct
			if entry.Tag == dwarf.TagMember {

				currentField := struct_field{}

				for _, field := range entry.Field {

					if field.Attr == dwarf.AttrName {
						currentField.name = field.Val.(string)
					}

					if field.Attr == dwarf.AttrDataMemberLoc {
						currentField.offset = field.Val.(int64)
					}

					if field.Attr == dwarf.AttrType {

						typeReader.Seek(field.Val.(dwarf.Offset))
						typeEntry, err := typeReader.Next()
						if err != nil {
							return nil, err
						}

						for i := range typeEntry.Field {
							if typeEntry.Field[i].Attr == dwarf.AttrName {
								currentField.typeName = typeEntry.Field[i].Val.(string)
							}
						}

					}
				}
				currentlyReadingStruct.fields = append(currentlyReadingStruct.fields, currentField)
			}
		}
	}
	return structs, nil
}
