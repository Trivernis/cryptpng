# Cryptpng ![](https://img.shields.io/discord/729250668162056313)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FTrivernis%2Fcryptpng.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FTrivernis%2Fcryptpng?ref=badge_shield)


A way to store encrypted data inside a png without altering the image itself.

## Usage

```shell script
# encrypt
cryptpng encrypt --image <name of the image> --in <input file> --out <output file>

# decrypt
cryptpng decrypt --image <crypt image> --out <decrypted output file>
```

## Technical Information

It should be possible to store data with a size up to ~ 4GB, but in reality most image viewers have
problems with chunks that are bigger than several Megabytes.
The data itself is stored in a [png chunk](http://www.libpng.org/pub/png/spec/1.2/PNG-Structure.html)
and encrypted via aes. The encryption chunk is stored right before the `IDAT` chunk that contains the
image data. The steps for encrypting are:

### Encrypt

1. Parse the png file and split it into chunks.
2. Prompt for a password and use the scrypt 32byte value with a generated salt.
3. Store the salt in the `saLt` chunk.
4. Encrypt the data using aes and the provided hashed key.
5. Split the data into parts of 1 MiB of size.
6. Store every data part into a separate `crPt` chunk.
7. Write the png header and chunks to the output file.

### Decrypt

1. Parse the png file and split it into chunks.
2. Get the `saLt` chunk.
3. Get the `crPt` chunks and and concat the data.
4. Prompt for the password and create the scrypt 32byte hash with the salt.
5. Decrypt the data using aes and the provided hash key.
6. Write the data to the specified output file.


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FTrivernis%2Fcryptpng.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FTrivernis%2Fcryptpng?ref=badge_large)