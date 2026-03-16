# dart_mnemonicode

A Dart port of the original [mnemonicode](https://github.com/singpolyma/mnemonicode) project.

This package provides a method for encoding binary data into a sequence of words which can be spoken over the phone, for example, and converted back to data on the other side.

It uses a specific 1626-word list designed to have no soundalikes and to be easily recognizable by non-native English speakers.

**Note:** This is *not* the same as BIP39 used in cryptocurrency wallets. It is an older, general-purpose binary-to-text encoding.

## Installation

Add the following to your `pubspec.yaml`:

```yaml
dependencies:
  dart_mnemonicode: ^1.0.0
```

## Usage

### Public API

```dart
import 'package:dart_mnemonicode/dart_mnemonicode.dart';

void main() {
  // Encoding bytes to words
  var bytes = [0x12, 0x34, 0x56, 0x78];
  var encoded = encodeBytes(bytes);
  print(encoded); // e.g. "academy-acrobat-active"

  // Decoding words to bytes
  var decodedBytes = decodePhrase(encoded);
  print(decodedBytes);

  // Encoding a hex string
  var hexStr = "12345678";
  var encodedHex = encodeHex(hexStr);
  print(encodedHex);

  // Decoding words to a hex string
  var decodedHex = decodeToHex(encodedHex);
  print(decodedHex);
}
```

### CLI

The package includes two executables: `mnencode` and `mndecode`.

To encode from standard input to standard output:
```bash
echo -n "Hello" | dart run bin/mnencode.dart
```

To encode hex string input:
```bash
echo -n "48656c6c6f" | dart run bin/mnencode.dart -x
```

To decode from standard input to standard output:
```bash
echo -n "academy acrobat active" | dart run bin/mndecode.dart
```

To decode to hex string output:
```bash
echo -n "academy acrobat active" | dart run bin/mndecode.dart -x
```

## Compatibility Notes

This Dart implementation aims to be bit-for-bit compatible with the original C implementation, retaining the exact same wordlist, semantics, and output formatting.
