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
2. Prompt for a password and use the sha512 32byte value.
3. Create a base64 string out of the data.
4. Encrypt the base64 string using aes and the provided hashed key.
5. Store the data into the `crPt` chunk.
6. Write the png header and chunks to the output file.

### Decrypt

1. Parse the png file and split it into chunks.
2. Get the `crPt` chunk.
3. Prompt for the password and create the sha512 32byte hash.
4. Decrypt the data using aes and the provided hash key.
5. Decode the base64 data.
6. Write the data to the specified output file.