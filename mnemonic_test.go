package mnemonic_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"mnemonic"
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
