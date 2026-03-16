package mnemonic_test

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/arran4/go-mnemonicode"
)

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
	}{
		{"empty", []byte{}},
		{"hello", []byte("hello")},
		{"world", []byte("world!")},
		{"hex1", []byte{0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56}},
		{"hex2", []byte{0xFF, 0x00, 0xAA, 0x55}},
		{"long", []byte("This is a much longer string to test the encoding and decoding capabilities of the mnemonic library.")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := mnemonic.Encode(tt.in, "")
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			dec, err := mnemonic.Decode(enc)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if !bytes.Equal(tt.in, dec) {
				t.Errorf("expected %v, got %v", tt.in, dec)
			}
		})
	}
}

func TestStreamEncodeDecode(t *testing.T) {
	in := []byte("Streaming test data to ensure the io.Writer and io.Reader implementations work properly.")

	var buf bytes.Buffer
	enc := mnemonic.NewEncoder(&buf, "")
	_, err := enc.Write(in[:10])
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	_, err = enc.Write(in[10:])
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	err = enc.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	dec := mnemonic.NewDecoder(&buf)
	out, err := io.ReadAll(dec)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if !bytes.Equal(in, out) {
		t.Errorf("expected %s, got %s", in, out)
	}
}

func FuzzEncodeDecode(f *testing.F) {
	f.Add([]byte("hello"))
	f.Add([]byte(""))
	f.Add([]byte{0x00, 0xFF, 0xAA})

	f.Fuzz(func(t *testing.T, data []byte) {
		enc, err := mnemonic.Encode(data, "")
		if err != nil {
			t.Fatalf("Encode failed: %v", err)
		}

		dec, err := mnemonic.Decode(enc)
		if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}

		if !bytes.Equal(data, dec) {
			t.Errorf("Roundtrip failed. Expected %v, got %v", data, dec)
		}
	})
}

func BenchmarkEncode(b *testing.B) {
	data := []byte(strings.Repeat("0123456789abcdef", 100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mnemonic.Encode(data, "")
	}
}

func BenchmarkDecode(b *testing.B) {
	data := []byte(strings.Repeat("0123456789abcdef", 100))
	enc, _ := mnemonic.Encode(data, "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mnemonic.Decode(enc)
	}
}

// TestGoldenVectors ensures the implementation stays backwards compatible
// with the output from the original C version and other known good outputs.
func TestGoldenVectors(t *testing.T) {
	// Format is expected to be "x x x. x x x. x x x\n" from the C tools
	// Note: Our Decode implementation accepts any standard mnemonic format regardless of separation.
	tests := []struct {
		name     string
		hexInput string // the hex input
		expected string // the expected decoded mnemonic
	}{
		{
			name:     "hello",
			hexInput: "68656c6c6f", // "hello"
			expected: "square angel stone. carlo",
		},
		{
			name:     "hex_zeros",
			hexInput: "00000000",
			expected: "academy academy academy",
		},
		{
			name:     "mixed_bytes",
			hexInput: "0102030405",
			expected: "papa twist alpine. admiral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert hex string to byte array
			var in []byte
			for i := 0; i < len(tt.hexInput); i += 2 {
				val, _ := strconv.ParseUint(tt.hexInput[i:i+2], 16, 8)
				in = append(in, byte(val))
			}

			enc, err := mnemonic.Encode(in, "x x x. x x x. x x x\n")
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			// The original mnencode appended a newline, but the library Encode only outputs the format
			// as strictly requested. For our test we trim it.
			enc = strings.TrimSpace(enc)

			if enc != tt.expected {
				t.Errorf("Expected encoded '%s', got '%s'", tt.expected, enc)
			}

			dec, err := mnemonic.Decode(enc)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if !bytes.Equal(in, dec) {
				t.Errorf("Expected decoded '%x', got '%x'", in, dec)
			}
		})
	}
}

func TestBadFormat(t *testing.T) {
	in := []byte("hello")
	_, err := mnemonic.Encode(in, "xxx") // Bad format, missing separator
	if err == nil {
		t.Errorf("Expected error for bad format, got nil")
	}
}
