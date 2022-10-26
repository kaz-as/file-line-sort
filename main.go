package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pbnjay/memory"
)

type Arguments struct {
	InputFilename        string
	OutputFilename       string
	MaxBytesMemoryForUse uint
}

func parseInputArguments() Arguments {
	args := Arguments{}

	flag.StringVar(&args.OutputFilename, "o", "", "output filename")
	flag.StringVar(&args.InputFilename, "i", "", "input filename")
	flag.UintVar(&args.MaxBytesMemoryForUse, "m", 0, "max memory size in bytes for program to use")

	flag.Parse()

	return args
}

func checkInputArguments(args Arguments) error {
	if args.InputFilename == "" {
		return fmt.Errorf("input file should be specified")
	}

	if args.OutputFilename == "" {
		return fmt.Errorf("output file should be specified")
	}

	if free := memory.FreeMemory(); uint64(args.MaxBytesMemoryForUse) > free {
		return fmt.Errorf("max memory for use %d cannot be larger than free memory %d",
			args.MaxBytesMemoryForUse,
			free,
		)
	}

	return nil
}

func main() {
	args := parseInputArguments()

	if err := checkInputArguments(args); err != nil {
		fmt.Printf("input arguments error: %s", err)
		os.Exit(1)
	}

	sorter := FileSorter{
		In:             args.InputFilename,
		Out:            args.OutputFilename,
		MaxBytesMemory: args.MaxBytesMemoryForUse,
	}

	if err := sorter.Sort(); err != nil {
		fmt.Printf("sort error: %s", err)
		os.Exit(1)
	}
}
