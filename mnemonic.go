// Package mnemonic implements a method for encoding binary data into a sequence
// of words which can be spoken over the phone, for example, and converted
// back to data on the other side.
package mnemonic

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const (
	MnBase       = 1626                 // cubic root of 2^32, rounded up
	MnRemainder  = 7                    // extra words for 24 bit remainders
	MnWords      = MnBase + MnRemainder // total number of words
	MnWordBufLen = 25                   // size for a word buffer+headroom

	MnEof = 0 // signal end to mn_decode_word_index

	// result codes for C-like API
	MnOk         = 0
	MnErem       = -1
	MnEoverrun   = -2
	MnEoverrun24 = -3
	MnEindex     = -4
	MnEindex24   = -5
	MnEencoding  = -6
	MnEword      = -7
	MnEformat    = -8

	MnFdefault        = "x-x-x--"
	MnF64BitsPerLine  = " x-x-x--x-x-x\n"
	MnF96BitsPerLine  = " x-x-x--x-x-x--x-x-x\n"
	MnF128BitsPerLine = " x-x-x--x-x-x--x-x-x--x-x-x\n"
)

// Idiomatic Go errors
var (
	ErrUnexpectedRemainder = errors.New("mnemonic: unexpected arithmetic remainder")
	ErrBufferOverrun       = errors.New("mnemonic: output buffer overrun")
	ErrOverrun24           = errors.New("mnemonic: data after 24 bit remainder")
	ErrBadWordIndex        = errors.New("mnemonic: bad word index")
	ErrUnexpected24BitRem  = errors.New("mnemonic: unexpected 24 bit remainder word")
	ErrInvalidEncoding     = errors.New("mnemonic: invalid arithmetic encoding")
	ErrUnrecognizedWord    = errors.New("mnemonic: unrecognized word")
	ErrBadFormat           = errors.New("mnemonic: bad format string")
)

// CodeToError maps C-like error codes to Go errors.
func CodeToError(code int) error {
	switch code {
	case MnOk:
		return nil
	case MnErem:
		return ErrUnexpectedRemainder
	case MnEoverrun:
		return ErrBufferOverrun
	case MnEoverrun24:
		return ErrOverrun24
	case MnEindex:
		return ErrBadWordIndex
	case MnEindex24:
		return ErrUnexpected24BitRem
	case MnEencoding:
		return ErrInvalidEncoding
	case MnEword:
		return ErrUnrecognizedWord
	case MnEformat:
		return ErrBadFormat
	default:
		return fmt.Errorf("mnemonic: unknown error code %d", code)
	}
}

// mnWordsRequired returns the number of words required to encode data
func mnWordsRequired(size int) int {
	return ((size + 1) * 3) / 4
}

// MnEncodeWordIndex performs one step of encoding binary data into words. Returns word index.
func MnEncodeWordIndex(src []byte, srcsize int, n int) uint {
	var x uint32 = 0
	var offset int
	var remaining int
	var extra int = 0
	var i int

	if n < 0 || n >= mnWordsRequired(srcsize) {
		return 0 // word out of range
	}
	offset = (n / 3) * 4
	remaining = srcsize - offset
	if remaining <= 0 {
		return 0
	}
	if remaining >= 4 {
		remaining = 4
	}
	for i = 0; i < remaining; i++ {
		x |= uint32(src[offset+i]) << (i * 8)
	}

	switch n % 3 {
	case 2: // Third word of group
		if remaining == 3 { // special case for 24 bits
			extra = MnBase // use one of the 7 3-letter words
		}
		x /= (MnBase * MnBase)
	case 1: // Second word of group
		x /= MnBase
	}
	return uint((x % MnBase) + uint32(extra) + 1)
}

// MnEncodeWord performs one step of encoding binary data into words. Returns word.
func MnEncodeWord(src []byte, srcsize int, n int) string {
	idx := MnEncodeWordIndex(src, srcsize, n)
	if idx == 0 || int(idx) >= len(mnWords) {
		return ""
	}
	return mnWords[idx]
}

// MnNextWordIndex performs one step of decoding a string into word indices.
func MnNextWordIndex(ptr *string) uint {
	s := *ptr
	var wordbuf bytes.Buffer
	i := 0

	// skip separator chars
	for i < len(s) && !unicode.IsLetter(rune(s[i])) {
		i++
	}
	s = s[i:]
	wordstart := s

	i = 0
	for i < len(s) && unicode.IsLetter(rune(s[i])) && wordbuf.Len() < MnWordBufLen-1 {
		c := s[i]
		i++
		wordbuf.WriteByte(byte(unicode.ToLower(rune(c))))
	}
	s = s[i:]

	// skip tail of long words
	i = 0
	for i < len(s) && unicode.IsLetter(rune(s[i])) {
		i++
	}
	s = s[i:]

	// skip separators
	i = 0
	for i < len(s) && !unicode.IsLetter(rune(s[i])) {
		i++
	}
	s = s[i:]

	wordStr := wordbuf.String()
	if wordStr == "" {
		*ptr = s
		return 0 // EOF, no word found
	}

	for idx := 1; idx <= MnWords; idx++ {
		if mnWords[idx] == wordStr {
			*ptr = s
			return uint(idx)
		}
	}

	*ptr = wordstart // Not found
	return 0
}

// MnDecodeWordIndex performs one step of decoding a sequence of words into binary data.
func MnDecodeWordIndex(index uint, dest []byte, destsize int, offset *int) int {
	var x uint32
	var groupofs int
	var i int

	if *offset < 0 {
		return *offset
	}

	if index > MnWords {
		*offset = MnEindex
		return *offset
	}

	if *offset > destsize {
		*offset = MnEoverrun
		return *offset
	}

	if index > MnBase && *offset%4 != 2 {
		*offset = MnEindex24
		return *offset
	}

	groupofs = *offset & ^3
	x = 0
	for i = 0; i < 4; i++ {
		if groupofs+i < destsize {
			x |= uint32(dest[groupofs+i]) << (i * 8)
		}
	}

	if index == MnEof {
		switch *offset % 4 {
		case 3:
			return MnOk
		case 2:
			if x <= 0xFFFF {
				return MnOk
			} else {
				*offset = MnErem
				return *offset
			}
		case 1:
			if x <= 0xFF {
				return MnOk
			} else {
				*offset = MnErem
				return *offset
			}
		case 0:
			return MnOk
		}
	}

	if *offset == destsize {
		*offset = MnEoverrun
		return *offset
	}

	index-- // 1 based to 0 based index

	switch *offset % 4 {
	case 3:
		*offset = MnEoverrun24
		return *offset
	case 2:
		if index >= MnBase {
			x += uint32(index-MnBase) * MnBase * MnBase
			(*offset)++
		} else {
			if index >= 1625 || (index == 1624 && x > 1312671) {
				*offset = MnEencoding
				return *offset
			}
			x += uint32(index) * MnBase * MnBase
			(*offset) += 2
		}
	case 1:
		x += uint32(index) * MnBase
		(*offset)++
	case 0:
		x = uint32(index)
		(*offset)++
	}

	for i = 0; i < 4; i++ {
		if groupofs+i < destsize {
			dest[groupofs+i] = byte(x % 256)
			x /= 256
		}
	}
	return MnOk
}

// MnEncode encodes a binary data buffer into a sequence of words.
func MnEncode(src []byte, srcsize int, dest []byte, destsize int, format string) int {
	if format == "" {
		format = MnFdefault
	}

	fmtStr := format
	fmtIdx := 0
	destIdx := 0

	for n := 0; n < mnWordsRequired(srcsize); n++ {
		for destIdx < destsize && fmtIdx < len(fmtStr) && !unicode.IsLetter(rune(fmtStr[fmtIdx])) {
			dest[destIdx] = fmtStr[fmtIdx]
			destIdx++
			fmtIdx++
		}

		if destIdx >= destsize {
			return MnEoverrun
		}

		if fmtIdx >= len(fmtStr) {
			if len(fmtStr) > 0 && unicode.IsLetter(rune(fmtStr[len(fmtStr)-1])) && unicode.IsLetter(rune(format[0])) {
				return MnEformat
			}
			fmtIdx = 0
			for destIdx < destsize && fmtIdx < len(fmtStr) && !unicode.IsLetter(rune(fmtStr[fmtIdx])) {
				dest[destIdx] = fmtStr[fmtIdx]
				destIdx++
				fmtIdx++
			}
			if fmtIdx < len(fmtStr) && !unicode.IsLetter(rune(fmtStr[fmtIdx])) {
				return MnEformat
			}
		}

		word := MnEncodeWord(src, srcsize, n)
		if word == "" {
			return MnEoverrun // shouldn't happen, actually
		}

		for fmtIdx < len(fmtStr) && unicode.IsLetter(rune(fmtStr[fmtIdx])) {
			fmtIdx++
		}

		wordIdx := 0
		for destIdx < destsize && wordIdx < len(word) {
			dest[destIdx] = word[wordIdx]
			destIdx++
			wordIdx++
		}
	}
	if destIdx < destsize {
		dest[destIdx] = 0 // null terminate
		// don't increment destIdx if we want to return MnOk but wait:
		// the previous code was "if destIdx < destsize { dest[destIdx] = 0; destIdx++ } else return MnEoverrun; return destIdx - 1"
		// The original C code returns MN_OK (0). Let's return MN_OK.
		dest[destIdx] = 0
	} else {
		return MnEoverrun
	}
	return MnOk
}

// MnDecode decodes a text representation in string src to binary buffer dest.
func MnDecode(src string, dest []byte, destsize int) int {
	var index uint
	offset := 0
	var status int

	ptr := src
	for {
		index = MnNextWordIndex(&ptr)
		if index == 0 {
			break
		}
		MnDecodeWordIndex(index, dest, destsize, &offset)
	}

	if ptr != "" {
		return MnEword
	}
	status = MnDecodeWordIndex(MnEof, dest, destsize, &offset)
	if status < 0 {
		return status
	}
	return offset
}

// ---- Idiomatic Go API ----

// Encode encodes a byte slice into a mnemonic string using the given format.
// If format is empty, MnFdefault is used.
func Encode(src []byte, format string) (string, error) {
	if format == "" {
		format = MnFdefault
	}


	// We'll reimplement it more idiomatically and efficiently:
	var sb strings.Builder
	fmtStr := format
	fmtIdx := 0

	for n := 0; n < mnWordsRequired(len(src)); n++ {
		for fmtIdx < len(fmtStr) && !unicode.IsLetter(rune(fmtStr[fmtIdx])) {
			sb.WriteByte(fmtStr[fmtIdx])
			fmtIdx++
		}

		if fmtIdx >= len(fmtStr) {
			if len(fmtStr) > 0 && unicode.IsLetter(rune(fmtStr[len(fmtStr)-1])) && unicode.IsLetter(rune(format[0])) {
				return "", ErrBadFormat
			}
			fmtIdx = 0
			for fmtIdx < len(fmtStr) && !unicode.IsLetter(rune(fmtStr[fmtIdx])) {
				sb.WriteByte(fmtStr[fmtIdx])
				fmtIdx++
			}
			if fmtIdx < len(fmtStr) && !unicode.IsLetter(rune(fmtStr[fmtIdx])) {
				return "", ErrBadFormat
			}
		}

		word := MnEncodeWord(src, len(src), n)
		if word == "" {
			return "", ErrBufferOverrun // Or a more appropriate error
		}

		for fmtIdx < len(fmtStr) && unicode.IsLetter(rune(fmtStr[fmtIdx])) {
			fmtIdx++
		}

		sb.WriteString(word)
	}
	return sb.String(), nil
}

// Decode decodes a mnemonic string into a byte slice.
func Decode(src string) ([]byte, error) {
	// Better yet, just count words.
	ptr := src
	wordCount := 0
	for {
		idx := MnNextWordIndex(&ptr)
		if idx == 0 {
			break
		}
		wordCount++
	}

	if ptr != "" {
		return nil, ErrUnrecognizedWord
	}

	maxSize := ((wordCount * 4) / 3) + 4
	dest := make([]byte, maxSize)

	n := MnDecode(src, dest, len(dest))
	if n < 0 {
		return nil, CodeToError(n)
	}
	return dest[:n], nil
}

// ---- io.Reader / io.Writer interfaces ----

// Encoder wraps an io.Writer and encodes data written to it into mnemonic format.
// Since mnemonic encoding requires knowing the total size up front to correctly
// compute words required and 24-bit remainder cases, we must buffer the entire
// stream in memory before encoding it.
type Encoder struct {
	w      io.Writer
	format string
	buf    bytes.Buffer
}

// NewEncoder returns a new Encoder that writes to w using the given format.
func NewEncoder(w io.Writer, format string) *Encoder {
	if format == "" {
		format = MnFdefault
	}
	return &Encoder{w: w, format: format}
}

// Write buffers the data. It does not write to the underlying writer until Close is called.
func (e *Encoder) Write(p []byte) (n int, err error) {
	return e.buf.Write(p)
}

// Close encodes the buffered data and writes it to the underlying writer.
func (e *Encoder) Close() error {
	encoded, err := Encode(e.buf.Bytes(), e.format)
	if err != nil {
		return err
	}
	_, err = io.WriteString(e.w, encoded)
	return err
}

// Decoder wraps an io.Reader and decodes mnemonic data read from it.
type Decoder struct {
	r          io.Reader
	decoded    []byte
	decodedIdx int
}

// NewDecoder returns a new Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Read reads decoded data into p.
func (d *Decoder) Read(p []byte) (n int, err error) {
	if d.decoded == nil {
		// Read all data from underlying reader
		data, err := io.ReadAll(d.r)
		if err != nil {
			return 0, err
		}
		decoded, err := Decode(string(data))
		if err != nil {
			return 0, err
		}
		d.decoded = decoded
		d.decodedIdx = 0
	}

	if d.decodedIdx >= len(d.decoded) {
		return 0, io.EOF
	}

	n = copy(p, d.decoded[d.decodedIdx:])
	d.decodedIdx += n
	return n, nil
}
