import 'package:test/test.dart';
import 'package:dart_mnemonicode/dart_mnemonicode.dart';

void main() {
  group('Mnemonicode API', () {
    test('encodeBytes and decodePhrase', () {
      final bytes = [0x54, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x61, 0x20, 0x74, 0x65, 0x73, 0x74, 0x20, 0x6f, 0x66, 0x20, 0x68, 0x65, 0x78, 0x20, 0x65, 0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67];
      final encoded = encodeBytes(bytes);
      final decoded = decodePhrase(encoded);

      expect(decoded.toList(), equals(bytes));
    });

    test('encodeHex and decodeToHex', () {
      final hexString = '5468697320697320612074657374206f662068657820656e636f64696e67';
      final encoded = encodeHex(hexString);
      final decoded = decodeToHex(encoded);

      expect(decoded.toLowerCase(), equals(hexString.toLowerCase()));
    });

    test('invalid input handling', () {
      expect(() => decodePhrase('invalid_word_that_is_not_in_list'), throwsException);
    });
  });
}
