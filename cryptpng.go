package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
)

type ChunkData struct {
	length uint32
	name string
	data []byte
	crc uint32
	raw []byte
}

var filename string
func main() {
	flag.StringVar(&filename, "file", "image.png", "The path of the png file.")
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	valid, header := validatePng(f)
	if valid {
		fout, err := os.Create("encrypted-" + filename)
		if err != nil {
			log.Fatal(err)
		}
		defer fout.Close()
		_, _ = fout.Write(header)
		chunk, err := readChunk(f)
		if err != nil {
			log.Fatal(err)
		}
		for chunk.name != "IDAT" {
			fmt.Printf("l: %d, n: %s, c: %d\n", chunk.length,  chunk.name, chunk.crc)
			_, _ = fout.Write(chunk.raw)
			chunk, err = readChunk(f)
			if err != nil {
				log.Fatal(err)
			}
		}
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Data to encrypt: ")
		inputDataString, _ := reader.ReadString('\n')
		inputData := []byte(inputDataString)
		inputData, err = encryptData(inputData)
		if err != nil {
			panic(err)
		}
		cryptChunk := createChunk(inputData, "crPt")
		_, _ = fout.Write(cryptChunk.raw)
		for {
			_, _ = fout.Write(chunk.raw)
			chunk, err = readChunk(f)
			if err != nil {
				break
			}
		}
	} else {
		log.Fatal("Invalid png.")
	}
}

// validates the png by reading the header of the file
func validatePng(f *os.File) (bool, []byte) {
	headerBytes := make([]byte, 8)
	_, err := f.Read(headerBytes)
	if err != nil {
		log.Fatal(err)
	}
	firstByteMatch := headerBytes[0] == 0x89
	pngAsciiMatch := string(headerBytes[1:4]) == "PNG"
	dosCRLF := headerBytes[4] == 0x0d && headerBytes[5] == 0x0a
	dosEof := headerBytes[6] == 0x1a
	unixLF := headerBytes[7] == 0x0a
	return firstByteMatch && pngAsciiMatch && dosCRLF && dosEof && unixLF, headerBytes
}

// reads the data of one chunk
// it is assumed that the file reader is at the beginning of the chunk when reading
func readChunk(f *os.File) (ChunkData, error) {
	lengthRaw := make([]byte, 4)
	_, err := f.Read(lengthRaw)
	length := binary.BigEndian.Uint32(lengthRaw)
	crcRaw := make([]byte, 4)
	nameRaw := make([]byte, 4)
	_, _ = f.Read(nameRaw)
	name := string(nameRaw)
	data := make([]byte, length)
	_, err = f.Read(data)
	_, err = f.Read(crcRaw)
	crc := binary.BigEndian.Uint32(crcRaw)
	fullData := make([]byte, 0)
	fullData = append(fullData, lengthRaw...)
	fullData = append(fullData, nameRaw...)
	fullData = append(fullData, data...)
	fullData = append(fullData, crcRaw...)
	return ChunkData{
		length: length,
		name:   name,
		data:   data,
		crc:    crc,
		raw: 	fullData,
	}, err
}

// creates a chunk with the given data and name
func createChunk(data []byte, name string) ChunkData {
	rawLength := make([]byte, 4)
	binary.BigEndian.PutUint32(rawLength, uint32(len(data)))
	rawName := []byte(name)
	dataAndName := make([]byte, 0)
	dataAndName = append(dataAndName, rawName...)
	dataAndName = append(dataAndName, data...)
	crc := crc32.ChecksumIEEE(dataAndName)
	rawCrc := make([]byte, 4)
	binary.BigEndian.PutUint32(rawCrc, crc)
	fullData := make([]byte, 0)
	fullData = append(fullData, rawLength...)
	fullData = append(fullData, dataAndName...)
	fullData = append(fullData, rawCrc...)
	return ChunkData{
		length: uint32(len(data)),
		name:   name,
		data:   data,
		crc:    crc,
		raw: 	fullData,
	}
}

// creates an encrypted png chunk
func encryptData(data []byte) ([]byte, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Password: ")
	pw, _ := reader.ReadString('\n')
	key := make([]byte, 32 - len(pw))
	key = append(key, []byte(pw)...)
	return encrypt(key, data)
}

func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	cipherText := make([]byte, aes.BlockSize+len(b))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], []byte(b))
	return cipherText, nil
}