package gotir

import (
	"debug/dwarf"
	"debug/elf"
	"io"
)

// ParseFromPath reads in all of the type information from the DWARF section of the ELF at the given patho
func ParseFromPath(path string) (*Gotir, error) {
	elfFile, err := elf.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := elfFile.DWARF()
	if err != nil {
		return nil, err
	}

	ir, err := parseFromData(data)
	if err != nil {
		return nil, err
	}

	return ir, nil
}

func parseFromData(data *dwarf.Data) (*Gotir, error) {

	lineReader := data.Reader()
	typeReader := data.Reader()

	ir := &Gotir{
		Structs:   []*struct_type{},
		Functions: []*function_type{},
	}

	var currentlyReadingStruct *struct_type = nil
	var currentlyReadingFunction *function_type = nil

entryReadLoop:
	for {
		entry, err := lineReader.Next()
		if err == io.EOF || entry == nil { //FIXME: Is `|| entry == nil` correct?
			break
		}
		if err != nil {
			return nil, err
		}
		if entryIsNull(entry) {
			if currentlyReadingStruct != nil {
				ir.Structs = append(ir.Structs, currentlyReadingStruct)
				currentlyReadingStruct = nil
			}
			if currentlyReadingFunction != nil {
				ir.Functions = append(ir.Functions, currentlyReadingFunction)
				currentlyReadingFunction = nil
			}
			continue entryReadLoop
		}

		// Found a struct
		if entry.Tag == dwarf.TagStructType {
			currentlyReadingStruct = readStructInit(entry)
		}
		// If currently reading the fields of a struct
		if currentlyReadingStruct != nil {
			err = readStructField(typeReader, entry, currentlyReadingStruct)
			if err != nil {
				return nil, err
			}
		}

		// Found a function
		if entry.Tag == dwarf.TagSubprogram {
			currentlyReadingFunction = readFunctionInit(entry)
		}
		// If currently reading the parameters of a function
		if currentlyReadingFunction != nil && entry.Tag == dwarf.TagFormalParameter {
			err = readFunctionParameter(typeReader, entry, currentlyReadingFunction)
			if err != nil {
				return nil, err
			}
		}
	}

	return ir, nil
}

func readFunctionInit(entry *dwarf.Entry) *function_type {
	currentlyReadingFunction := &function_type{}

	for _, field := range entry.Field {
		if field.Attr == dwarf.AttrName {
			currentlyReadingFunction.Name = field.Val.(string)
		}
	}

	currentlyReadingFunction.Params = []function_param{}
	return currentlyReadingFunction
}

func readFunctionParameter(typeReader *dwarf.Reader, entry *dwarf.Entry, currentlyReadingFunction *function_type) error {

	var (
		typeEntry *dwarf.Entry
		err       error
	)

	newParam := function_param{Name: "_", TypeName: "_"}
	for _, field := range entry.Field {

		if field.Attr == dwarf.AttrName {
			newParam.Name = field.Val.(string)
		}

		if field.Attr == dwarf.AttrVarParam {
			newParam.IsReturn = true
		}

		// Get the name of the type of the parameter
		// XXX: Have to go back later to get the size
		if field.Attr == dwarf.AttrType {
			typeReader.Seek(field.Val.(dwarf.Offset))
			typeEntry, err = typeReader.Next()
			if err != nil {
				return err
			}
			for i := range typeEntry.Field {
				if typeEntry.Field[i].Attr == dwarf.AttrName {
					newParam.TypeName = typeEntry.Field[i].Val.(string)
				}
			}
		}
	}

	currentlyReadingFunction.Params = append(currentlyReadingFunction.Params, newParam)

	return nil
}

func readStructInit(entry *dwarf.Entry) *struct_type {

	currentlyReadingStruct := &struct_type{}

	for _, field := range entry.Field {

		if field.Attr == dwarf.AttrName {
			currentlyReadingStruct.Name = field.Val.(string)
		}

		if field.Attr == dwarf.AttrByteSize {
			currentlyReadingStruct.Size = field.Val.(int64)
		}
	}

	currentlyReadingStruct.Fields = []struct_field{}
	return currentlyReadingStruct
}

func readStructField(typeReader *dwarf.Reader, entry *dwarf.Entry, currentlyReadingStruct *struct_type) error {

	if entry.Tag != dwarf.TagMember {
		return nil
	}

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
				return err
			}

			for i := range typeEntry.Field {
				if typeEntry.Field[i].Attr == dwarf.AttrName {
					currentField.TypeName = typeEntry.Field[i].Val.(string)
				}
			}
		}
	}
	currentlyReadingStruct.Fields = append(currentlyReadingStruct.Fields, currentField)
	return nil
}

func entryIsNull(e *dwarf.Entry) bool {
	return e.Children == false &&
		len(e.Field) == 0 &&
		e.Offset == 0 &&
		e.Tag == dwarf.Tag(0)
}
