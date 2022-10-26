package main

import (
	"flag"
	"fmt"
	"os"
)

type Arguments struct {
	Folder      string
	Count       uint
	MinLineSize uint
	MaxLineSize uint
	MinFileSize uint64
	MaxFileSize uint64
}

func parseInputArguments() Arguments {
	args := Arguments{}

	flag.StringVar(&args.Folder, "i", "", "input filename")
	flag.UintVar(&args.Count, "c", 0, "how many files to generate")
	flag.UintVar(&args.MinLineSize, "minline", 0, "min line size")
	flag.UintVar(&args.MaxLineSize, "maxline", 0, "max line size")
	flag.Uint64Var(&args.MinFileSize, "minfile", 0, "min file size")
	flag.Uint64Var(&args.MaxFileSize, "maxfile", 0, "max file size")

	flag.Parse()

	return args
}

func checkInputArguments(args Arguments) error {
	if args.Folder == "" {
		return fmt.Errorf("folder should be specified")
	}

	if args.MinLineSize > args.MaxLineSize {
		return fmt.Errorf("min line size should not be greater than max line size")
	}

	if args.MinFileSize > args.MaxFileSize {
		return fmt.Errorf("min file size should not be greater than max file size")
	}

	return nil
}

func main() {
	args := parseInputArguments()

	if err := checkInputArguments(args); err != nil {
		fmt.Printf("input arguments error: %s", err)
		os.Exit(1)
	}

	// todo
}
