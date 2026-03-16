import 'package:dart_mnemonicode/dart_mnemonicode.dart';

void main() {
  // 1. Encoding a byte array
  var bytes = [0x12, 0x34, 0x56, 0x78];
  var encodedPhrase = encodeBytes(bytes);
  print('Encoded bytes $bytes to: $encodedPhrase');

  // 2. Decoding the phrase back to bytes
  var decodedBytes = decodePhrase(encodedPhrase);
  print('Decoded phrase "$encodedPhrase" to: $decodedBytes');

  // 3. Encoding a hex string
  var hexStr = 'deadbeef';
  var encodedHex = encodeHex(hexStr);
  print('Encoded hex "$hexStr" to: $encodedHex');

  // 4. Decoding the phrase back to a hex string
  var decodedHex = decodeToHex(encodedHex);
  print('Decoded phrase "$encodedHex" to: $decodedHex');
}
