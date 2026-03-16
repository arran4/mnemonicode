package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/arran4/go-mnemonicode"
)

// EncodeCmd is a subcommand `go-mnemonic encode`
// EncodeCmd encodes binary data from stdin into a mnemonic word sequence on stdout.
//
// Flags:
//
//	hexEncodedInput: -x Read input as hex encoded string instead of raw bytes
//	format: -f (default: "x x x. x x x. x x x\n") Output format string
func EncodeCmd(hexEncodedInput bool, format string) error {
	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading stdin: %v", err)
	}

	if hexEncodedInput {
		inputData, err = hex2bytes(inputData)
		if err != nil {
			return fmt.Errorf("error parsing hex input: %v", err)
		}
	}

	encoded, err := mnemonic.Encode(inputData, format)
	if err != nil {
		return fmt.Errorf("error encoding data: %v", err)
	}

	fmt.Print(encoded)
	if encoded != "" && encoded[len(encoded)-1] != '\n' {
		fmt.Println()
	}
	return nil
}

// DecodeCmd is a subcommand `go-mnemonic decode`
// DecodeCmd decodes a mnemonic word sequence from stdin into binary data on stdout.
//
// Flags:
//
//	hexEncodedOutput: -x Output hex encoded string instead of raw bytes
func DecodeCmd(hexEncodedOutput bool) error {
	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading stdin: %v", err)
	}

	decoded, err := mnemonic.Decode(string(inputData))
	if err != nil {
		return fmt.Errorf("error decoding data: %v", err)
	}

	if hexEncodedOutput {
		for _, b := range decoded {
			fmt.Printf("%02X", b)
		}
		fmt.Println()
	} else {
		_, err := os.Stdout.Write(decoded)
		if err != nil {
			return fmt.Errorf("error writing output: %v", err)
		}
	}

	return nil
}

func isxdigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func counthex(s []byte) int {
	total := 0
	for _, b := range s {
		if isxdigit(b) {
			total++
		}
	}
	return total
}

func countLeadingCrapAndZeros(in []byte) int {
	for i, b := range in {
		if b != '0' && isxdigit(b) {
			return i // Find first non-zero hex digit
		}
	}
	return len(in)
}

func hex2bytes(in []byte) ([]byte, error) {
	var out []byte
	var t [3]byte
	j := 0

	i := countLeadingCrapAndZeros(in)
	in = in[i:]

	if counthex(in)%2 != 0 {
		t[j] = '0'
		j++
	}

	for _, b := range in {
		if isxdigit(b) {
			t[j] = b
			j++
			if j == 2 {
				val, err := strconv.ParseUint(string(t[:2]), 16, 8)
				if err != nil {
					return nil, err
				}
				out = append(out, byte(val))
				j = 0
			}
		}
	}
	return out, nil
}
