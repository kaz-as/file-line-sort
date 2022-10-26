package main

import (
	"fmt"
	"os"
)

type FileSorter struct {
	In             string
	Out            string
	MaxBytesMemory uint
}

const MaxByteCountToCopy = 50000000

func (s FileSorter) logSort(in []byte) ([]string, uint) {
	// todo написать сортировку n log n
	return []string{}, 0
}

func (s FileSorter) insertToSorted(in []string, add string) []string {
	// todo написать вставку в отсортированный слайс
	return []string{}
}

func (s FileSorter) Sort() error {
	in, err := os.OpenFile(s.In, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("open input file error: %s", err)
	}
	defer in.Close()

	out, err := os.OpenFile(s.Out, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open output file error: %s", err)
	}
	defer out.Close()

	fullOffset, _ := in.Seek(0, 2)
	lastByte := []byte{'\n'}
	readLast, _ := in.Read(lastByte)
	if readLast != 1 {
		return fmt.Errorf("cannot read last symbol of input file")
	}
	var outSize int64
	if lastByte[0] == '\n' {
		outSize = fullOffset + 1
	} else {
		outSize = fullOffset + 2
	}
	if err := out.Truncate(outSize); err != nil {
		return fmt.Errorf("change output file size error: %s", err)
	}
	_, _ = in.Seek(0, 0)

	maxSingleArraySize := int(s.MaxBytesMemory / 4)

	forLogSort := make([]byte, maxSingleArraySize)

	// В рамках сортировки считается, что слово содержит в себе перенос строки

	var left uint       // длина обрубленного в конце слова, которое нужно считать заново
	var sorted []string // быстро отсортированная часть считанного буфера
	var read int        // считано в последний раз

	var (
		sortedOldLeftPos  int64 = 0
		sortedOldRightPos int64 = 0
		processingLeft    int64 = 0
		processingRight   int64 = 0
	)

	bytesToCopy := make([]byte, MaxByteCountToCopy)

	for {
		read, err = in.Read(forLogSort)
		if err != nil {
			return fmt.Errorf("read input file error: %s", err)
		}
		if read == 0 || read == maxSingleArraySize {
			break
		}

		sorted, left = s.logSort(forLogSort[:read])

		currentSumReadStringLen := uint(read) - left

		if left == uint(read) {
			if read == maxSingleArraySize {
				return fmt.Errorf("too large line encountered, size=%d", left)
			}

			last := forLogSort[uint(len(forLogSort))-left:]
			if last[len(last)-1] != '\n' {
				last = append(last, '\n')
			}

			sorted = s.insertToSorted(sorted, string(last))
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
				sortedOldLeftPos = sortedOldRightPos - MaxByteCountToCopy
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

				sortedLessOrEqualThanCurrentOld := false
				for !sortedLessOrEqualThanCurrentOld {
					for _ = startOfWord; startOfWord < int(sortedOldRightPos-sortedOldLeftPos) && bytesToCopy[startOfWord] != '\n'; startOfWord++ {
					}
					startOfWord++

					if startOfWord >= int(sortedOldRightPos-sortedOldLeftPos) {
						break
					}

					sortedLessOrEqualThanCurrentOld = true
					currString := sorted[currChunkPos]
					for i := 0; i < len(currString); i++ {
						if bytesToCopy[i+startOfWord] < currString[i] {
							sortedLessOrEqualThanCurrentOld = false
							needNextToLeft = false
							break
						}

						if bytesToCopy[i+startOfWord] == '\n' {
							break
						}
					}
				}

				if sortedLessOrEqualThanCurrentOld {
					lenToCopy := sortedOldRightPos - sortedOldLeftPos - int64(startOfWord)
					processingLeft -= lenToCopy
					_, _ = out.Seek(processingLeft, 0)
					_, err = out.Write(bytesToCopy[startOfWord:])
					if err != nil {
						return fmt.Errorf("writing error: %v", err)
					}
				}
				sortedOldLeftPos += int64(startOfWord)
			}

			_, err = out.WriteString(sorted[currChunkPos])
			if err != nil {
				return fmt.Errorf("writing error: %v", err)
			}
		}
	}

	return nil
}
