package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/pbnjay/memory"
)

type Arguments struct {
	InputFilename  string
	OutputFilename string

	MaxBytesMemoryForUse uint64
}

func parseInputArguments() Arguments {
	args := Arguments{}

	flag.StringVar(&args.OutputFilename, "o", "", "output filename")
	flag.StringVar(&args.InputFilename, "i", "", "input filename")

	flag.Uint64Var(&args.MaxBytesMemoryForUse, "m", 0, "approx. max memory size for program to use")

	flag.Parse()

	return args
}

func checkInputArguments(args Arguments) error {
	if args.InputFilename == "" {
		return fmt.Errorf("input file must be specified")
	}

	if args.OutputFilename == "" {
		return fmt.Errorf("output file must be specified")
	}

	if free := memory.FreeMemory(); args.MaxBytesMemoryForUse > free {
		return fmt.Errorf("max memory for use %d cannot be larger than free memory %d",
			args.MaxBytesMemoryForUse,
			free,
		)
	}

	return nil
}

func prepareInputArguments(args *Arguments) {
	if args.MaxBytesMemoryForUse == 0 {
		maxAllowed := memory.FreeMemory() / 3 * 2
		if maxAllowed > math.MaxUint {
			maxAllowed = math.MaxUint
		}
		args.MaxBytesMemoryForUse = maxAllowed
	}
}

func main() {
	args := parseInputArguments()

	if err := checkInputArguments(args); err != nil {
		fmt.Printf("input arguments error: %s", err)
		os.Exit(1)
	}

	prepareInputArguments(&args)

	sorter := FileSorter{
		In:             args.InputFilename,
		Out:            args.OutputFilename,
		MaxBytesMemory: args.MaxBytesMemoryForUse,
	}

	t1 := time.Now()
	if err := sorter.Sort(); err != nil {
		fmt.Printf("sort error: %s", err)
		os.Exit(1)
	}
	fmt.Printf("minutes spent: %v", time.Now().Sub(t1).Minutes())
}
