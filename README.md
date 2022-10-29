# File line sorting program

## Overview

The program sorts lines of large files without extra memory usage more than O(1).

The program does not change input file and does not create any temporary files.

## Description

The program contains a sorting part itself, a large file generator for testing purposes,
and a program for checking correctness of sorting.

For simplicity the program is guaranteed to work for files that contain only English alphabet and spaces.

## Usage

How to run programs after build to `bin` folder.

### Sort
#### Parameters:
Required:
* `-i` - an input file
* `-o` - an output file

Optional:
* `-m` - approximate max memory size for program to use
* `-mc` - copy buffer size, must be greater than max line size

#### Example
```bash
./bin/sort -i /path/to/folder/input_4d65822107fcfd52.txt -o /path/to/folder/sorted_4d65822107fcfd52.txt -m 100000000 -mc 30000000
```

### Generate
#### Parameters:
Required:
* `-i` - a folder where to create a generated file
* `-l` - max line length
* `-s` - line count

Optional:
* `-c` - number of files to generate
* `-prefix` - prefix of a filename
* `-suffix` - suffix of a filename

#### Example
```bash
./bin/generate -i /path/to/folder/ -c 1 -l 500 -s 800000 -prefix input_ -suffix .txt
```

200 Mb file is processed per 4 hours with the program with the parameters above.

### Checker
All arguments are files to check.

If a file is not sorted, the program prints info about that.

#### Example
```bash
./bin/check /path/to/folder/file1.txt /path/to/folder/file2.txt /path/to/another/folder/file.txt
```

## Algorithm
Asymptotic is O(n^2).

Due to forbiddance of any temporary files, the program cannot use external sorting algorithms, thus asymptotic
cannot be improved.

Sort O(n log n) chunks of `-m` size. For a sorted chunk in memory and sorted all previous data in output file search do
for each line in the chunk starting from the end find a chunk in the file of `-mc` size which contains the "rightest"
line greater or equal than the line in memory *(O(n) for each, can be improved to O(log n))* and move all data in the
output file from this found position to the right to the final (for this sorted chunk in memory) position and write the
line in the sorted chunk right before the moved data.

## Further improvements
1. Make search of an element in already sorted part of the output file not linear, but binary.
It would improve constant, not asymptotic.
2. If usage of temporary files is allowed, use standard external sorting algorithms.
It would improve asymptotic to O(n log n).
