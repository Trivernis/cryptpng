# Cryptpng

A proof of concept implementation of storing encrypted data inside of png metadata chunks.

## Usage

```shell script
# encrypt
cryptpng --image <name of the image> --in <input file> --out <output file>

# decrypt
cryptpng --decrypt --image <crypt image> --out <decrypted output file>
```