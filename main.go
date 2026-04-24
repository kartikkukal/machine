package main

import (
	"flag"
	"fmt"
	"machine/file"
	"machine/processor"
	"os"
)

var (
	execPath = flag.String("exec", "main", "path of the executable file")
)

func main() {

	flag.Parse()

	bytes, err := os.ReadFile(*execPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	parser := file.NewParser(bytes)
	exec, err := parser.Parse()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	state := processor.New(exec)

	err = state.Execute()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
