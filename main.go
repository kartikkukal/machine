package main

import (
	"flag"
	"fmt"
	"machine/processor"
	"os"
)

var (
	execPath = flag.String("exec", "main", "path of the executable file")
)

func main() {

	flag.Parse()

	exec, err := os.ReadFile(*execPath)
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
