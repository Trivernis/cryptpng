# Cryptpng

A proof of concept implementation of storing encrypted data inside of png metadata chunks.

## Usage

```shell script
# encrypt
cryptpng --image <name of the image> --in <input file> --out <output file>

# decrypt
cryptpng --decrypt --image <crypt image> --out <decrypted output file>
```

## Technical Information

It should be possible to store data with a size up to ~ 4GB, but in reality most image viewers have
problems with chunks that are bigger than several Megabytes.
The data itself is stored in a [png chunk](http://www.libpng.org/pub/png/spec/1.2/PNG-Structure.html)
and encrypted via aes. The encryption chunk is stored right before the `IDAT` chunk that contains the
image data. The steps for encrypting are:

### Encrypt

1. Parse the png file and split it into chunks.
2. Prompt for a password and use the sha512 32byte value with a generated salt.
3. Store the salt in the `saLt` chunk.
4. Create a base64 string out of the data.
5. Encrypt the base64 string using aes and the provided hashed key.
6. Split the data into parts of 1 MiB of size.
7. Store every data part into a separate `crPt` chunk.
8. Write the png header and chunks to the output file.

### Decrypt

1. Parse the png file and split it into chunks.
2. Get the `saLt` chunk.
3. Get the `crPt` chunks and and concat the data.
4. Prompt for the password and create the sha512 32byte hash with the salt.
5. Decrypt the data using aes and the provided hash key.
6. Decode the base64 data.
7. Write the data to the specified output file.