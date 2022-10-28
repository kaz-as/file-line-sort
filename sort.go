package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

type FileSorter struct {
	In             string
	Out            string
	MaxBytesMemory uint
	MaxBytesBuffer uint
}

const MaxRealCountOfExistedLargeArrays = 2
const Separator = '\n'

type byteSlices [][]byte

func (b byteSlices) Len() int {
	return len(b)
}

func (b byteSlices) Less(i, j int) bool {
	return bytes.Compare(b[i], b[j]) == -1
}

func (b byteSlices) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// logSort сортировка чанка за n log n
func (s FileSorter) logSort(in []byte) (sorted [][]byte, left uint) {
	start := 0
	curr := 0

	for ; curr < len(in); curr++ {
		if in[curr] == Separator {
			sorted = append(sorted, in[start:curr+1])
			start = curr + 1
		}
	}

	left = uint(len(in) - start)

	sort.Sort(byteSlices(sorted))

	return
}

func (s FileSorter) getMaxSingleArraySize() int {
	return int(s.MaxBytesMemory / MaxRealCountOfExistedLargeArrays)
}

func (s FileSorter) insertToSorted(in [][]byte, add []byte) (out [][]byte) {
	out = append(in, add)
	var i int
	defer func() {
		out[i] = add
	}()

	for i = len(in) - 1; i > 0; i-- {
		if bytes.Compare(in[i-1], add) < 1 {
			return
		}
		in[i] = in[i-1]
	}
	return
}

func (s FileSorter) Sort() error {
	in, err := os.OpenFile(s.In, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("open input file error: %s", err)
	}
	defer in.Close()

	out, err := os.OpenFile(s.Out, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open output file error: %s", err)
	}
	defer out.Close()

	fullOffset, err := in.Seek(-1, 2)
	if err != nil {
		return fmt.Errorf("seek to last byte error: %s", err)
	}
	lastByte := []byte{Separator}
	readLast, err := in.Read(lastByte)
	if err != nil {
		return fmt.Errorf("read last byte error: %s", err)
	}
	if readLast != 1 {
		return fmt.Errorf("cannot read last symbol of input file")
	}
	var outSize int64
	if lastByte[0] == Separator {
		outSize = fullOffset + 1
	} else {
		outSize = fullOffset + 2
	}
	if err := out.Truncate(outSize); err != nil {
		return fmt.Errorf("change output file size error: %s", err)
	}
	_, _ = in.Seek(0, 0)

	maxSingleArraySize := s.getMaxSingleArraySize()

	forLogSort := make([]byte, maxSingleArraySize)

	// В рамках сортировки считается, что слово содержит в себе перенос строки

	var left uint       // длина обрубленного в конце слова, которое нужно считать заново
	var sorted [][]byte // быстро отсортированная часть считанного буфера
	var read int        // считано в последний раз

	var (
		sortedOldLeftPos  int64 = 0
		sortedOldRightPos int64 = 0
		processingLeft    int64 = 0
		processingRight   int64 = 0
	)

	bytesToCopy := make([]byte, s.MaxBytesBuffer)

	for {
		read, err = in.Read(forLogSort)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read input file error: %s", err)
		}
		if read == 0 {
			break
		}

		sorted, left = s.logSort(forLogSort[:read])

		currentSumReadStringLen := uint(read) - left

		if left == uint(read) {
			if read == maxSingleArraySize {
				return fmt.Errorf("too large line encountered, size=%d", left)
			}

			last := forLogSort[uint(len(forLogSort))-left:]
			if last[len(last)-1] != Separator {
				last = []byte(string(last) + string(Separator))
			}

			sorted = s.insertToSorted(sorted, last)
			currentSumReadStringLen += uint(len(last))
		} else {
			_, err = in.Seek(int64(-left), 1)
			if err != nil {
				return fmt.Errorf("change offset in input file error: %s", err)
			}
		}

		// вставка в файл вывода с последовательным сдвигом с конца уже записанного

		sortedOldRightPos = processingRight

		processingRight += int64(currentSumReadStringLen)
		processingLeft = processingRight

		for currChunkPos := len(sorted) - 1; currChunkPos >= 0; currChunkPos-- {
			needNextToLeft := true // делать false, когда надо переходить на следующую позицию в новом отсортированном

			for needNextToLeft {
				sortedOldLeftPos = sortedOldRightPos - int64(s.MaxBytesBuffer)
				if sortedOldLeftPos < 0 {
					sortedOldLeftPos = 0
				}
				if sortedOldLeftPos == 0 {
					needNextToLeft = false
				}

				_, _ = out.Seek(sortedOldLeftPos, 0)
				_, err = out.Read(bytesToCopy[:sortedOldRightPos-sortedOldLeftPos])
				if err != nil {
					return fmt.Errorf("read output file error: %s", err)
				}

				startOfWord := 0

				if sortedOldLeftPos != 0 {
					for ; startOfWord < int(sortedOldRightPos-sortedOldLeftPos) && bytesToCopy[startOfWord] != Separator; startOfWord++ {
					}
					startOfWord++
				}

				sortedLessOrEqualThanCurrentOld := false
				for !sortedLessOrEqualThanCurrentOld && startOfWord < int(sortedOldRightPos-sortedOldLeftPos) {
					sortedLessOrEqualThanCurrentOld = true
					currString := sorted[currChunkPos]
					for i := 0; i < len(currString); i++ {
						if bytesToCopy[i+startOfWord] < currString[i] {
							sortedLessOrEqualThanCurrentOld = false
							needNextToLeft = false
							break
						}
						if bytesToCopy[i+startOfWord] > currString[i] {
							break
						}

						if bytesToCopy[i+startOfWord] == Separator || currString[i] == Separator {
							break
						}
					}

					if !sortedLessOrEqualThanCurrentOld {
						for ; startOfWord < int(sortedOldRightPos-sortedOldLeftPos) && bytesToCopy[startOfWord] != Separator; startOfWord++ {
						}
						startOfWord++
					}
				}

				if sortedLessOrEqualThanCurrentOld {
					lenToCopy := sortedOldRightPos - sortedOldLeftPos - int64(startOfWord)
					processingLeft -= lenToCopy
					_, _ = out.Seek(processingLeft, 0)
					_, err = out.Write(bytesToCopy[startOfWord : sortedOldRightPos-sortedOldLeftPos])
					if err != nil {
						return fmt.Errorf("writing error: %v", err)
					}
				}
				if sortedOldRightPos > sortedOldLeftPos {
					sortedOldRightPos = sortedOldLeftPos + int64(startOfWord)
				}
			}

			processingLeft -= int64(len(sorted[currChunkPos]))
			_, _ = out.Seek(processingLeft, 0)
			_, err = out.Write(sorted[currChunkPos])
			if err != nil {
				return fmt.Errorf("writing error: %v", err)
			}
		}
	}

	return nil
}
