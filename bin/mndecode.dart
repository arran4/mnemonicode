import 'dart:io';
import 'dart:typed_data';
import 'package:dart_mnemonicode/src/mnemonic.dart';

void main(List<String> args) {
  bool hexEncodedOutput = false;

  if (args.isNotEmpty && args[0] == "-x") {
    hexEncodedOutput = true;
  }

  BytesBuilder builder = BytesBuilder();
  int byte;
  while ((byte = stdin.readByteSync()) != -1) {
    builder.addByte(byte);
  }
  String buf = String.fromCharCodes(builder.toBytes());

  Uint8List outbuf = Uint8List(0x10000);

  int n = mn_decode(buf, outbuf, outbuf.length);
  if (n < 0) {
    stderr.writeln("mn_decode result \$n");
    exit(1);
  }

  if (hexEncodedOutput) {
    for (int i = 0; i < n; i++) {
      stdout.write(outbuf[i].toRadixString(16).padLeft(2, '0').toUpperCase());
    }
  } else {
    stdout.add(Uint8List.sublistView(outbuf, 0, n));
  }
}
