package main

import (
	"fmt"
	"machine/state"
	"os"
)

func main() {

	image, err := os.ReadFile("a.v")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	/*for i, b := range image {
		fmt.Printf("%08b ", b)

		if (i+1)%8 == 0 {
			fmt.Println()
		}
	}*/

	//fmt.Println(len(image))

	state := state.New(image)
	err = state.Execute()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
