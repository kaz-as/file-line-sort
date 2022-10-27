package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
)

type Arguments struct {
	Folder      string
	MaxLineSize int

	Prefix   string
	Suffix   string
	Count    uint
	FileSize uint64
}

func parseInputArguments() Arguments {
	args := Arguments{}

	flag.StringVar(&args.Folder, "i", "", "input folder (must exist)")
	flag.IntVar(&args.MaxLineSize, "l", 1_000_000, "max line size")

	flag.StringVar(&args.Prefix, "prefix", "", "additional: output filename prefix")
	flag.StringVar(&args.Suffix, "suffix", "", "additional: output filename suffix")
	flag.UintVar(&args.Count, "c", 1, "additional: how many files to generate")
	flag.Uint64Var(&args.FileSize, "s", 30_000_000_000, "additional: file size")

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

	fileSizeFinishCondition := args.FileSize - uint64(args.MaxLineSize) - 1
	alphabet := getAlphabet()

	for i := uint(0); i < args.Count; i++ {
		err := func() error {
			newFileName := args.Prefix + strconv.FormatUint(rand.Uint64(), 16) + args.Suffix
			file, err := os.Create(filepath.Join(args.Folder, newFileName))
			if err != nil {
				return fmt.Errorf("cannot create file %s: %s", newFileName, err)
			}
			defer file.Close()

			w := bufio.NewWriter(file)

			for currFileSize := uint64(0); currFileSize < fileSizeFinishCondition; {
				currLineSize := rand.Intn(args.MaxLineSize + 1)
				var line []byte
				for j := 0; j < currLineSize; j++ {
					line = append(line, alphabet[rand.Intn(len(alphabet))])
				}
				line = append(line, '\n')
				_, err := w.Write(line)
				if err != nil {
					return err
				}

				currFileSize += uint64(currLineSize)
			}

			return w.Flush()
		}()

		if err != nil {
			fmt.Printf("error while processing %d file: %s", i, err)
			os.Exit(1)
		}
	}
}
