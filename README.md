# File line sorting program

## Overview

The program sorts lines of large files using the external sort algorithm.

## Description

The program contains a sorting part itself, a large-file generator for testing purposes,
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

#### Example
```bash
./bin/sort -i /path/to/folder/input_4d65822107fcfd52.txt -o /path/to/folder/sorted_4d65822107fcfd52.txt -m 100000000
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

The checker does not compare the sets of strings from initial and resulting files.

#### Example
```bash
./bin/check /path/to/folder/file1.txt /path/to/folder/file2.txt /path/to/another/folder/file.txt
```
