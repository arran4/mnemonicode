import sys
import re

def convert_wordlist():
    with open("mn_wordlist.c", "r") as f:
        content = f.read()

    # The array declaration looks like `const char *mn_words[MN_WORDS + 1] = { 0, ...`
    match = re.search(r'const char \*mn_words\[MN_WORDS \+ 1\] = {\s*0,\s*(.*?)};', content, re.DOTALL)
    if not match:
        print("Could not find mn_words array")
        return

    array_content = match.group(1)
    words = re.findall(r'"([^"]+)"', array_content)

    version_match = re.search(r'const char \*mn_wordlist_version =\s*"([^"]+)";', content)
    version = version_match.group(1) if version_match else "Wordlist ver 0.7"

    go_code = f"""// TODO: port complete
package mnemonic

// WordlistVersion is the version of the mnemonic wordlist
const WordlistVersion = "{version}"

// mnWords is the mnemonic wordlist
var mnWords = []string{{
\t"",
"""
    for i in range(0, len(words), 6):
        batch = words[i:i+6]
        formatted_batch = ", ".join(f'"{w}"' for w in batch)
        go_code += f"\t{formatted_batch},\n"
    go_code += "}\n"

    with open("wordlist.go", "w") as f:
        f.write(go_code)

if __name__ == "__main__":
    convert_wordlist()
