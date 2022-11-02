package main

import (
	"bufio"
	"bytes"
	"container/heap"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
)

type FileSorter struct {
	In             string
	Out            string
	MaxBytesMemory uint64
}

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
func (s FileSorter) logSort(in []byte) (sorted [][]byte, left int) {
	start := 0
	curr := 0

	for ; curr < len(in); curr++ {
		if in[curr] == Separator {
			sorted = append(sorted, in[start:curr+1])
			start = curr + 1
		}
	}

	left = len(in) - start

	sort.Sort(byteSlices(sorted))

	return
}

func (s FileSorter) maxLineSize() (uint64, error) {
	in, err := os.OpenFile(s.In, os.O_RDONLY, 0)
	if err != nil {
		return 0, fmt.Errorf("open input file error: %w", err)
	}
	defer in.Close()

	size := math.MaxInt
	if s.MaxBytesMemory < uint64(size) {
		size = int(s.MaxBytesMemory)
	}
	f := bufio.NewReaderSize(in, size)

	maxSize := uint64(0)
	currSize := uint64(0)
	b, err := f.ReadByte()
	for err == nil {
		if b == Separator {
			if maxSize < currSize {
				maxSize = currSize
			}
			currSize = 0
		} else {
			currSize++
		}

		b, err = f.ReadByte()
	}

	if errors.Is(err, io.EOF) {
		if maxSize < currSize {
			maxSize = currSize
		}
		return maxSize, nil
	}

	return 0, err
}

func (s FileSorter) Sort() error {
	//maxLineSize, err := s.maxLineSize()
	//if err != nil {
	//	return fmt.Errorf("get max line size error: %w", err)
	//}
	oneMemory := math.MaxInt
	if s.MaxBytesMemory/3 < uint64(oneMemory) {
		oneMemory = int(s.MaxBytesMemory / 3)
	}
	//if maxLineSize > uint64(oneMemory/20) {
	//	return fmt.Errorf("too large string exists len=%d", maxLineSize)
	//}

	in, err := os.OpenFile(s.In, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("open input file error: %w", err)
	}
	defer in.Close()

	inStat, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stat input file error: %w", err)
	}

	out, err := os.OpenFile(s.Out, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open output file error: %w", err)
	}
	defer out.Close()
	err = out.Truncate(inStat.Size())
	if err != nil {
		return fmt.Errorf("change output file size error: %w", err)
	}

	tempDir, err := ioutil.TempDir("", "file-line-sort-")
	if err != nil {
		return fmt.Errorf("create temp directory error: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempFilesApproxCountInt64 := inStat.Size() / int64(oneMemory)
	if tempFilesApproxCountInt64 >= math.MaxInt {
		return fmt.Errorf("too small RAM amount allowed to use")
	}
	tempFilesApproxCountInt64 = tempFilesApproxCountInt64 + tempFilesApproxCountInt64/50 + 1
	if tempFilesApproxCountInt64 >= math.MaxInt {
		return fmt.Errorf("too small RAM amount allowed to use")
	}

	tempFilesApproxCount := int(tempFilesApproxCountInt64)

	tempFiles := make([]*os.File, 0, tempFilesApproxCount)
	defer func() {
		for _, file := range tempFiles {
			file.Close()
		}
	}()

	// первичная сортировка
	inputBuffer := make([]byte, oneMemory)
	for {
		n, err := in.Read(inputBuffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error reading input file: %w", err)
		}
		realInputBuffer := inputBuffer[:n]

		sorted, left := s.logSort(realInputBuffer)
		_, err = in.Seek(int64(-left), 1)
		if err != nil {
			return fmt.Errorf("seek input file error: %w", err)
		}

		file, err := ioutil.TempFile(tempDir, "tmpfile-*")
		if err != nil {
			return err
		}
		err = file.Truncate(int64(n - left))
		if err != nil {
			return fmt.Errorf("change temp file size error: %w", err)
		}

		tmpBuf := bufio.NewWriterSize(file, oneMemory)
		for _, str := range sorted {
			_, err = tmpBuf.Write(str)
			if err != nil {
				return fmt.Errorf("write to temp file buffer error: %w", err)
			}
		}
		err = tmpBuf.Flush()
		if err != nil {
			return fmt.Errorf("flush to temp file error: %w", err)
		}

		tempFiles = append(tempFiles, file)
	}

	// n-merge
	outBuf := bufio.NewWriterSize(out, oneMemory/tempFilesApproxCount)
	inBufs := make([]*bufio.Reader, 0, len(tempFiles))
	for _, file := range tempFiles {
		_, err = file.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("tmp file seek error: %w", err)
		}
		inBufs = append(inBufs, bufio.NewReaderSize(file, oneMemory/tempFilesApproxCount))
	}

	hList := make([]heapElement, 0, len(tempFiles))
	for _, inBuf := range inBufs {
		str, err := inBuf.ReadString(Separator)
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			return fmt.Errorf("read from temp file error: %w", err)
		}

		hList = append(hList, heapElement{
			s:      str,
			reader: inBuf,
		})
	}

	h := (*heapList)(&hList)
	heap.Init(h)

	for h.Len() > 0 {
		min := (*h)[0]
		str, err := min.reader.ReadString(Separator)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("read from temp file error: %w", err)
			}

			toOut := heap.Pop(h).(heapElement).s
			_, err = outBuf.WriteString(toOut)
			if err != nil {
				return fmt.Errorf("write to output buffer error: %w", err)
			}

		} else {
			_, err = outBuf.WriteString(min.s)
			if err != nil {
				return fmt.Errorf("write to output buffer error: %w", err)
			}

			(*h)[0].s = str
			heap.Fix(h, 0)
		}
	}

	err = outBuf.Flush()
	if err != nil {
		return fmt.Errorf("flush output buffer error: %w", err)
	}

	return nil
}

type heapElement struct {
	s      string
	reader *bufio.Reader
}

type heapList []heapElement

func (h *heapList) Len() int {
	return len(*h)
}

func (h *heapList) Less(i, j int) bool {
	return (*h)[i].s < (*h)[j].s
}

func (h *heapList) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *heapList) Push(x any) {
	*h = append(*h, x.(heapElement))
}

func (h *heapList) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
