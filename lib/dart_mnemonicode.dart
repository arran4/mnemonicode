/// A Dart port of the original mnemonicode project, providing a method for encoding binary data into a sequence of easily spoken words.
library dart_mnemonicode;

import 'dart:typed_data';
import 'src/mnemonic.dart' as mn;

/// Encodes a list of bytes into a mnemonic phrase.
///
/// Uses the default format 'x-x-x--'.
String encodeBytes(List<int> bytes) {
  var src = Uint8List.fromList(bytes);
  var dest = StringBuffer();
  var result = mn.mn_encode(src, src.length, dest, 0x90000, null);
  if (result != mn.MN_OK) {
    throw Exception('mn_encode error $result');
  }
  return dest.toString();
}

/// Encodes a string of hex characters into a mnemonic phrase.
String encodeHex(String hexString) {
  var bytes = _hexToBytes(hexString);
  return encodeBytes(bytes);
}

/// Decodes a mnemonic phrase into a list of bytes.
Uint8List decodePhrase(String phrase) {
  var dest = Uint8List(0x10000);
  var result = mn.mn_decode(phrase, dest, dest.length);
  if (result < 0) {
    throw Exception('mn_decode error $result');
  }
  return Uint8List.sublistView(dest, 0, result);
}

/// Decodes a mnemonic phrase into a string of hex characters.
String decodeToHex(String phrase) {
  var bytes = decodePhrase(phrase);
  var buffer = StringBuffer();
  for (var b in bytes) {
    buffer.write(b.toRadixString(16).padLeft(2, '0').toUpperCase());
  }
  return buffer.toString();
}

Uint8List _hexToBytes(String hexStr) {
  var cleanHex = hexStr.replaceAll(RegExp(r'[^0-9A-Fa-f]'), '');
  if (cleanHex.length % 2 != 0) {
    cleanHex = '0$cleanHex';
  }
  var bytes = Uint8List(cleanHex.length ~/ 2);
  for (var i = 0; i < bytes.length; i++) {
    bytes[i] = int.parse(cleanHex.substring(i * 2, i * 2 + 2), radix: 16);
  }
  return bytes;
}
