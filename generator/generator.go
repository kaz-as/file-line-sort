package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Arguments struct {
	Folder      string
	MaxLineSize int
	LineCount   uint64

	Prefix string
	Suffix string
	Count  uint
}

func parseInputArguments() Arguments {
	args := Arguments{}

	flag.StringVar(&args.Folder, "i", "", "input folder (must exist)")
	flag.IntVar(&args.MaxLineSize, "l", 1_000_000, "max line size")
	flag.Uint64Var(&args.LineCount, "s", 20_000, "line count")

	flag.StringVar(&args.Prefix, "prefix", "", "additional: output filename prefix")
	flag.StringVar(&args.Suffix, "suffix", "", "additional: output filename suffix")
	flag.UintVar(&args.Count, "c", 1, "additional: how many files to generate")

	flag.Parse()

	return args
}

func checkInputArguments(args Arguments) error {
	if args.Folder == "" {
		return fmt.Errorf("folder must be specified")
	}

	return nil
}

func getAlphabet() (alpha []byte) {
	alpha = []byte{' '}
	for c := byte('a'); c <= 'z'; c++ {
		alpha = append(alpha, c)
	}
	for c := byte('A'); c <= 'Z'; c++ {
		alpha = append(alpha, c)
	}
	return
}

func main() {
	args := parseInputArguments()

	if err := checkInputArguments(args); err != nil {
		fmt.Printf("input arguments error: %s", err)
		os.Exit(1)
	}

	alphabet := getAlphabet()
	rnd := rand.New(rand.NewSource(time.Now().UnixMilli()))

	line := make([]byte, 0, args.MaxLineSize+1)
	for i := uint(0); i < args.Count; i++ {
		err := func() error {
			newFileName := args.Prefix + strconv.FormatUint(rnd.Uint64(), 16) + args.Suffix
			file, err := os.Create(filepath.Join(args.Folder, newFileName))
			if err != nil {
				return fmt.Errorf("cannot create file %s: %s", newFileName, err)
			}
			defer file.Close()

			w := bufio.NewWriter(file)

			for j := uint64(0); j < args.LineCount; j++ {
				currLineSize := rand.Intn(args.MaxLineSize + 1)
				line = line[:0]
				for k := 0; k < currLineSize; k++ {
					line = append(line, alphabet[rand.Intn(len(alphabet))])
				}
				line = append(line, '\n')
				_, err := w.Write(line)
				if err != nil {
					return err
				}
			}

			return w.Flush()
		}()

		if err != nil {
			fmt.Printf("error while processing %d file: %s", i, err)
			os.Exit(1)
		}
	}
}
