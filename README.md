# mnemonicode

A Go implementation of the mnemonic encoding method.

These routines implement a method for encoding binary data into a sequence
of words which can be spoken over the phone, for example, and converted
back to data on the other side.

For more information see <http://web.archive.org/web/20101031205747/http://www.tothink.com/mnemonic/>

## Installation

```sh
go get github.com/singpolyma/mnemonicode
```

## Usage

### Using the API

```go
package main

import (
	"fmt"
	"log"

	"github.com/singpolyma/mnemonicode" // Note: the module name might be just "mnemonic" locally
)

func main() {
	// Encoding
	data := []byte("hello")
	encoded, err := mnemonic.Encode(data, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Encoded:", encoded)

	// Decoding
	decoded, err := mnemonic.Decode(encoded)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Decoded:", string(decoded))
}
```

### Streaming API

You can also use the `io.Reader` and `io.Writer` interfaces:

```go
// Using NewEncoder and NewDecoder
// Note: mnemonic encoding requires chunking bytes in groups of 4 or knowing the full length
```

### Command Line Tools

You can install the command line tools:

```sh
go install github.com/singpolyma/mnemonicode/cmd/mnencode@latest
go install github.com/singpolyma/mnemonicode/cmd/mndecode@latest
```
