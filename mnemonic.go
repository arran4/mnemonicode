// TODO: port complete
package mnemonic

import (
	"bytes"
	"unicode"
)

const (
	MnBase       = 1626
	MnRemainder  = 7
	MnWords      = MnBase + MnRemainder
	MnWordBufLen = 25

	MnEof = 0

	// result codes
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
// Takes a pointer to a string, advances the pointer past the word it reads,
// and returns the index of the word.
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
		*ptr = s // actually, if we didn't find anything we update ptr? C says: "wordbuf[0] == '\0' return 0"
		return 0 // EOF, no word found
	}

	for idx := 1; idx <= MnWords; idx++ {
		if mnWords[idx] == wordStr {
			*ptr = s
			return uint(idx)
		}
	}

	*ptr = wordstart // Not found, point back to the beginning of unrecognized word
	return 0
}

// MnDecodeWordIndex performs one step of decoding a sequence of words into binary data.
func MnDecodeWordIndex(index uint, dest []byte, destsize int, offset *int) int {
	var x uint32
	var groupofs int
	var i int

	if *offset < 0 { // Error from previous call? report it
		return *offset
	}

	if index > MnWords { // Word index out of range
		*offset = MnEindex
		return *offset
	}

	if *offset > destsize { // out of range?
		*offset = MnEoverrun
		return *offset
	}

	if index > MnBase && *offset%4 != 2 {
		// Unexpected 24 bit remainder word
		*offset = MnEindex24
		return *offset
	}

	groupofs = *offset & ^3 // Offset of 4 byte group containing offset
	x = 0
	for i = 0; i < 4; i++ {
		if groupofs+i < destsize { // Ignore any bytes outside buffer
			x |= uint32(dest[groupofs+i]) << (i * 8) // assemble number
		}
	}

	if index == MnEof { // Got EOF signal
		switch *offset % 4 {
		case 3: // group was three words and the last word was a 24 bit remainder
			return MnOk
		case 2: // last group has two words
			if x <= 0xFFFF { // should encode 16 bit data
				return MnOk
			} else {
				*offset = MnErem
				return *offset
			}
		case 1: // last group has just one word
			if x <= 0xFF { // should encode 8 bits
				return MnOk
			} else {
				*offset = MnErem
				return *offset
			}
		case 0: // last group was full 3 words
			return MnOk
		}
	}

	if *offset == destsize { // At EOF but didn't get MN_EOF
		*offset = MnEoverrun
		return *offset
	}

	index-- // 1 based to 0 based index

	switch *offset % 4 {
	case 3: // Got data past 24 bit remainder
		*offset = MnEoverrun24
		return *offset
	case 2:
		if index >= MnBase {
			// 24 bit remainder
			x += uint32(index-MnBase) * MnBase * MnBase
			(*offset)++ // *offset%4 == 3 for next time
		} else {
			// catch invalid encodings
			if index >= 1625 || (index == 1624 && x > 1312671) {
				*offset = MnEencoding
				return *offset
			}
			x += uint32(index) * MnBase * MnBase
			(*offset) += 2 // *offset%4 == 0 for next time
		}
	case 1:
		x += uint32(index) * MnBase
		(*offset)++
	case 0:
		x = uint32(index)
		(*offset)++
	}

	for i = 0; i < 4; i++ {
		if groupofs+i < destsize { // Don't step outside the buffer
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
