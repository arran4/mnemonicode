# go-mnemonicode

A Go implementation of the mnemonic encoding method.

These routines implement a method for encoding binary data into a sequence
of words which can be spoken over the phone, for example, and converted
back to data on the other side.

This package provides a standalone, idiomatic Go library while preserving
exact algorithm and output compatibility with the original C implementation
of Mnemonicode.

For more information on Mnemonicode, see <http://web.archive.org/web/20101031205747/http://www.tothink.com/mnemonic/>

## Installation

```sh
go get github.com/arran4/go-mnemonicode
```

## Usage

### Using the API

```go
package main

import (
	"fmt"
	"log"

	"github.com/arran4/go-mnemonicode"
)

func main() {
	// Encoding
	data := []byte("hello")
	// An empty format string will use the default format: "x-x-x--"
	encoded, err := mnemonicode.Encode(data, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Encoded:", encoded)

	// Decoding
	decoded, err := mnemonicode.Decode(encoded)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Decoded:", string(decoded))
}
```

### Streaming API

You can also use the `io.Reader` and `io.Writer` interfaces:

```go
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/arran4/go-mnemonicode"
)

func main() {
	var buf bytes.Buffer
	enc := mnemonicode.NewEncoder(&buf, "")

	// Write data to the encoder
	enc.Write([]byte("hello "))
	enc.Write([]byte("world"))

	// Close the encoder to finalize the encoding
	if err := enc.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Encoded stream:", buf.String())

	// Read and decode the stream
	dec := mnemonicode.NewDecoder(&buf)
	decoded, err := io.ReadAll(dec)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Decoded stream:", string(decoded))
}
```

### Command Line Tools

You can install the command line tools:

```sh
go install github.com/arran4/go-mnemonicode/cmd/go-mnemonic@latest
```

**Usage:**

Encode data to mnemonic:
```sh
echo -n "hello" | go-mnemonic encode
```

Encode hex data:
```sh
echo -n "68656c6c6f" | go-mnemonic encode -x
```

Decode mnemonic:
```sh
echo -n "square-angel-stone--carlo" | go-mnemonic decode
```

Decode mnemonic to hex:
```sh
echo -n "square-angel-stone--carlo" | go-mnemonic decode -x
```
