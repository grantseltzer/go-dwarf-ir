package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	gotir "github.com/grantseltzer/dwarf-to-gotir/pkg"
)

func main() {
	ir, err := gotir.ParseFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.Marshal(ir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}
