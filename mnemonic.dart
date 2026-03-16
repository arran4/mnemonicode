import 'dart:typed_data';
import 'mn_wordlist.dart';

const int MN_BASE = 1626;
const int MN_REMAINDER = 7;
const int MN_WORDS = MN_BASE + MN_REMAINDER;
const int MN_WORD_BUFLEN = 25;

const int MN_EOF = 0;

const int MN_OK = 0;
const int MN_EREM = -1;
const int MN_EOVERRUN = -2;
const int MN_EOVERRUN24 = -3;
const int MN_EINDEX = -4;
const int MN_EINDEX24 = -5;
const int MN_EENCODING = -6;
const int MN_EWORD = -7;
const int MN_EFORMAT = -8;

const String MN_FDEFAULT = "x-x-x--";
const String MN_F64BITSPERLINE = " x-x-x--x-x-x\n";
const String MN_F96BITSPERLINE = " x-x-x--x-x-x--x-x-x\n";
const String MN_F128BITSPERLINE = " x-x-x--x-x-x--x-x-x--x-x-x\n";

int mn_words_required(int size) {
  return ((size + 1) * 3) ~/ 4;
}

int mn_encode_word_index(Uint8List src, int srcsize, int n) {
  int x = 0;
  int offset;
  int remaining;
  int extra = 0;
  int i;

  if (n < 0 || n >= mn_words_required(srcsize)) {
    return 0;
  }

  offset = (n ~/ 3) * 4;
  remaining = srcsize - offset;

  if (remaining <= 0) return 0;
  if (remaining >= 4) remaining = 4;

  for (i = 0; i < remaining; i++) {
    x |= src[offset + i] << (i * 8);
  }

  switch (n % 3) {
    case 2:
      if (remaining == 3) {
        extra = MN_BASE;
      }
      x ~/= (MN_BASE * MN_BASE);
      break;
    case 1:
      x ~/= MN_BASE;
  }

  return x % MN_BASE + extra + 1;
}

String? mn_encode_word(Uint8List src, int srcsize, int n) {
  int index = mn_encode_word_index(src, srcsize, n);
  if (index == 0) return null;
  return mn_words[index];
}

class NextWordResult {
  final int index;
  final int charsRead;
  NextWordResult(this.index, this.charsRead);
}

NextWordResult mn_next_word_index(String src, int offset) {
  int startOffset = offset;

  bool isAlpha(int codeUnit) {
    return (codeUnit >= 65 && codeUnit <= 90) || (codeUnit >= 97 && codeUnit <= 122);
  }

  while (offset < src.length && !isAlpha(src.codeUnitAt(offset))) {
    offset++;
  }

  int wordstart = offset;
  List<int> wordbuf = [];

  while (offset < src.length && isAlpha(src.codeUnitAt(offset)) && wordbuf.length < MN_WORD_BUFLEN - 1) {
    int c = src.codeUnitAt(offset++);
    if (c >= 65 && c <= 90) { // 'A' to 'Z'
      c += 32; // 'a' - 'A'
    }
    wordbuf.add(c);
  }

  String word = String.fromCharCodes(wordbuf);

  while (offset < src.length && isAlpha(src.codeUnitAt(offset))) {
    offset++;
  }
  while (offset < src.length && !isAlpha(src.codeUnitAt(offset))) {
    offset++;
  }

  if (word.isEmpty) {
    return NextWordResult(0, offset - startOffset);
  }

  for (int idx = 1; idx <= MN_WORDS; idx++) {
    if (word == mn_words[idx]) {
      return NextWordResult(idx, offset - startOffset);
    }
  }

  return NextWordResult(0, wordstart - startOffset);
}

int mn_decode_word_index(int index, Uint8List dest, int destsize, List<int> offsetRef) {
  int x;
  int groupofs;
  int i;
  int offset = offsetRef[0];

  if (offset < 0) {
    return offset;
  }

  if (index > MN_WORDS) {
    offsetRef[0] = MN_EINDEX;
    return MN_EINDEX;
  }

  if (offset > destsize) {
    offsetRef[0] = MN_EOVERRUN;
    return MN_EOVERRUN;
  }

  if (index > MN_BASE && offset % 4 != 2) {
    offsetRef[0] = MN_EINDEX24;
    return MN_EINDEX24;
  }

  groupofs = offset & ~3;
  x = 0;
  for (i = 0; i < 4; i++) {
    if (groupofs + i < destsize) {
      x |= dest[groupofs + i] << (i * 8);
    }
  }

  if (index == MN_EOF) {
    switch (offset % 4) {
      case 3:
        return MN_OK;
      case 2:
        if (x <= 0xFFFF) {
          return MN_OK;
        } else {
          offsetRef[0] = MN_EREM;
          return MN_EREM;
        }
      case 1:
        if (x <= 0xFF) {
          return MN_OK;
        } else {
          offsetRef[0] = MN_EREM;
          return MN_EREM;
        }
      case 0:
        return MN_OK;
    }
  }

  if (offset == destsize) {
    offsetRef[0] = MN_EOVERRUN;
    return MN_EOVERRUN;
  }

  index--;

  switch (offset % 4) {
    case 3:
      offsetRef[0] = MN_EOVERRUN24;
      return MN_EOVERRUN24;
    case 2:
      if (index >= MN_BASE) {
        x += (index - MN_BASE) * MN_BASE * MN_BASE;
        offset++;
      } else {
        if (index >= 1625 || (index == 1624 && x > 1312671)) {
          offsetRef[0] = MN_EENCODING;
          return MN_EENCODING;
        }
        x += index * MN_BASE * MN_BASE;
        offset += 2;
      }
      break;
    case 1:
      x += index * MN_BASE;
      offset++;
      break;
    case 0:
      x = index;
      offset++;
      break;
  }

  for (i = 0; i < 4; i++) {
    if (groupofs + i < destsize) {
      dest[groupofs + i] = x % 256;
      x ~/= 256;
    }
  }

  offsetRef[0] = offset;
  return MN_OK;
}

int mn_encode(Uint8List src, int srcsize, StringBuffer dest, int destsize, String? format) {
  int n;
  int fmtOffset = 0;
  String? word;

  bool isAlpha(int codeUnit) {
    return (codeUnit >= 65 && codeUnit <= 90) || (codeUnit >= 97 && codeUnit <= 122);
  }

  if (format == null || format.isEmpty) {
    format = MN_FDEFAULT;
  }

  for (n = 0; n < mn_words_required(srcsize); n++) {
    while (dest.length < destsize && fmtOffset < format.length && !isAlpha(format.codeUnitAt(fmtOffset))) {
      dest.writeCharCode(format.codeUnitAt(fmtOffset++));
    }

    if (dest.length >= destsize) {
      return MN_EOVERRUN;
    }

    if (fmtOffset >= format.length) {
      if (isAlpha(format.codeUnitAt(format.length - 1)) && isAlpha(format.codeUnitAt(0))) {
        return MN_EFORMAT;
      }
      fmtOffset = 0;
      while (dest.length < destsize && fmtOffset < format.length && !isAlpha(format.codeUnitAt(fmtOffset))) {
        dest.writeCharCode(format.codeUnitAt(fmtOffset++));
      }
      if (!isAlpha(format.codeUnitAt(fmtOffset))) {
        return MN_EFORMAT;
      }
    }

    word = mn_encode_word(src, srcsize, n);
    if (word == null) {
      return MN_EOVERRUN;
    }

    while (fmtOffset < format.length && isAlpha(format.codeUnitAt(fmtOffset))) {
      fmtOffset++;
    }

    for (int i = 0; i < word.length; i++) {
      if (dest.length < destsize) {
        dest.writeCharCode(word.codeUnitAt(i));
      }
    }
  }

  if (dest.length >= destsize) {
    return MN_EOVERRUN;
  }

  return MN_OK;
}

int mn_decode(String src, Uint8List dest, int destsize) {
  int srcOffset = 0;
  List<int> offset = [0];
  int status;

  while (true) {
    NextWordResult result = mn_next_word_index(src, srcOffset);
    if (result.index == 0) {
      srcOffset += result.charsRead;
      break;
    }
    mn_decode_word_index(result.index, dest, destsize, offset);
    srcOffset += result.charsRead;
  }

  if (srcOffset < src.length) {
    return MN_EWORD;
  }

  status = mn_decode_word_index(MN_EOF, dest, destsize, offset);
  if (status < 0) {
    return status;
  }

  return offset[0];
}
