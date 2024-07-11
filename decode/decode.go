package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
)

func main() {
	smallFile, err := os.Open("../binarytracks.bin")
	if err != nil {
		panic("cant open trackfile")
	}
	defer func() {
		if err := smallFile.Close(); err != nil {
			panic("cant close file")
		}
	}()
	data, err := io.ReadAll(smallFile)
	if err != nil {
		panic("Error leyendo archivo")
	}
	var arr []string
	smallBuffer := bytes.NewBuffer(data)
	smallDecoder := gob.NewDecoder(smallBuffer)
	if err := smallDecoder.Decode(&arr); err != nil {
		panic("can't decode")
	}
	fmt.Println(arr)
	fmt.Println(len(arr))
}
