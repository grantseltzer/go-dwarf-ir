package gotir

import (
	"debug/dwarf"
	"debug/elf"
	"io"
)

func ParseFromPath(path string) (*gotir, error) {
	elfFile, err := elf.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := elfFile.DWARF()
	if err != nil {
		return nil, err
	}

	structs, err := parseStructs(data)
	if err != nil {
		return nil, err
	}

	functions, err := parseFunctions(data)
	if err != nil {
		return nil, err
	}

	gotir := &gotir{
		Functions: functions,
		Structs:   structs,
	}

	return gotir, nil
}

func parseFunctions(data *dwarf.Data) ([]*function_type, error) {
	return nil, nil
}

func parseStructs(data *dwarf.Data) ([]*struct_type, error) {

	lineReader := data.Reader()
	typeReader := data.Reader()

	structs := []*struct_type{}
	var currentlyReadingStruct *struct_type = nil

entryReadLoop:
	for {
		entry, err := lineReader.Next()
		if err == io.EOF || entry == nil {
			break
		}
		if err != nil {
			return nil, err
		}
		if entryIsNull(entry) {
			if currentlyReadingStruct == nil {
				continue entryReadLoop
			}
			structs = append(structs, currentlyReadingStruct)
			currentlyReadingStruct = nil
		}

		// Found a struct
		if entry.Tag == dwarf.TagStructType {

			currentlyReadingStruct = &struct_type{}

			for _, field := range entry.Field {

				if field.Attr == dwarf.AttrName {
					currentlyReadingStruct.Name = field.Val.(string)
				}

				if field.Attr == dwarf.AttrByteSize {
					currentlyReadingStruct.Size = field.Val.(int64)
				}
			}

			currentlyReadingStruct.Fields = []struct_field{}
		}

		// If currently reading the Fields of a struct
		if currentlyReadingStruct != nil {

			// Found a member of the struct
			if entry.Tag == dwarf.TagMember {

				currentField := struct_field{}

				for _, field := range entry.Field {

					if field.Attr == dwarf.AttrName {
						currentField.Name = field.Val.(string)
					}

					if field.Attr == dwarf.AttrDataMemberLoc {
						currentField.Offset = field.Val.(int64)
					}

					if field.Attr == dwarf.AttrType {

						typeReader.Seek(field.Val.(dwarf.Offset))
						typeEntry, err := typeReader.Next()
						if err != nil {
							return nil, err
						}

						for i := range typeEntry.Field {
							if typeEntry.Field[i].Attr == dwarf.AttrName {
								currentField.TypeName = typeEntry.Field[i].Val.(string)
							}
						}

					}
				}
				currentlyReadingStruct.Fields = append(currentlyReadingStruct.Fields, currentField)
			}
		}
	}
	return structs, nil
}

func entryIsNull(e *dwarf.Entry) bool {
	return e.Children == false &&
		len(e.Field) == 0 &&
		e.Offset == 0 &&
		e.Tag == dwarf.Tag(0)
}
