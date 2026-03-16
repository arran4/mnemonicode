import 'dart:io';
import 'dart:typed_data';
import 'package:dart_mnemonicode/src/mnemonic.dart';
import 'package:dart_mnemonicode/src/mn_wordlist.dart';

bool isXDigit(int c) {
  return (c >= 48 && c <= 57) || // 0-9
         (c >= 65 && c <= 70) || // A-F
         (c >= 97 && c <= 102);  // a-f
}

int counthex(Uint8List s, int len) {
  int total = 0;
  for (int i = 0; i < len; i++) {
    if (isXDigit(s[i])) total++;
  }
  return total;
}

int countLeadingCrapAndZeros(Uint8List inBytes, int inlen) {
  int i;
  for (i = 0; i < inlen; i++) {
    if (inBytes[i] != 48 /* '0' */ && isXDigit(inBytes[i])) break;
  }
  return i;
}

int hex2bytes(int inlen, Uint8List inBytes, Uint8List out) {
  List<int> t = [0, 0];
  int j = 0;
  int k = 0;
  int offset = countLeadingCrapAndZeros(inBytes, inlen);

  inlen = inlen - offset;

  if (counthex(Uint8List.sublistView(inBytes, offset, offset + inlen), inlen) % 2 != 0) {
    t[j++] = 48; // '0'
  }

  for (int i = 0; i < inlen; i++) {
    if (isXDigit(inBytes[offset + i])) {
      t[j++] = inBytes[offset + i];
      if (j == 2) {
        String hexStr = String.fromCharCodes(t);
        int? val = int.tryParse(hexStr, radix: 16);
        if (val == null) {
          return -1;
        }
        out[k++] = val;
        j = 0;
      }
    }
  }
  return k;
}

void main(List<String> args) {
  bool hexEncodedInput = false;

  if (args.isNotEmpty && args[0] == "-x") {
    hexEncodedInput = true;
  }

  stderr.writeln(mn_wordlist_version);

  BytesBuilder builder = BytesBuilder();
  int byte;
  while ((byte = stdin.readByteSync()) != -1) {
    builder.addByte(byte);
  }

  Uint8List cbuf = builder.toBytes();
  int buflen = cbuf.length;

  Uint8List buf;
  if (hexEncodedInput) {
    Uint8List xbuf = Uint8List(0x8000);
    buflen = hex2bytes(buflen, cbuf, xbuf);
    if (buflen < 0) {
      stderr.writeln("hex2bytes error");
      exit(1);
    }
    buf = Uint8List.sublistView(xbuf, 0, buflen);
  } else {
    buf = cbuf;
  }

  StringBuffer outbuf = StringBuffer();
  int n = mn_encode(buf, buflen, outbuf, 0x90000, "x x x. x x x. x x x\n");

  if (n != 0) {
    stderr.writeln("mn_encode error \$n");
    exit(1);
  }

  stdout.write(outbuf.toString());
  stdout.write('\n');
}
