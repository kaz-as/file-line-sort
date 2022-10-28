package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"syscall"
)

func isSorted(in *bufio.Reader) bool {
	curr, err := in.ReadString('\n')
	if err == io.EOF {
		return true
	}
	if err != nil {
		return false
	}

	var prev string

	for {
		prev = curr
		curr, err = in.ReadString('\n')
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}

		if prev > curr {
			return false
		}
	}
}

func main() {
	var exit int
	for i := 1; i < len(os.Args); i++ {
		func() {
			filename := os.Args[i]
			file, err := os.Open(filename)
			if err != nil {
				panic(err)
			}
			defer file.Close()
			if !isSorted(bufio.NewReader(file)) {
				fmt.Printf("%s is not sorted", filename)
				exit = 1
			}
		}()
	}
	syscall.Exit(exit)
}
