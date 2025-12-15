# Implementation of Huffman Encoding and Decoding

This repository provides a complete implementation of Huffman encoding and decoding written in Go. The project focuses on correctness, performance, and robustness, and includes extensive unit testing, fuzz testing, and benchmarking to validate both functionality and efficiency.

The implementation supports encoding and decoding of arbitrary binary data and text, including non-ASCII input. Special attention is given to edge cases such as empty inputs, highly repetitive data, and inputs with a minimal number of unique symbols.

## Usage

### Encoding file
```bash
go-huffman -e input.txt
```

### Decoding file
```bash
go-huffman -d input.hfm -o input.res.txt
```
