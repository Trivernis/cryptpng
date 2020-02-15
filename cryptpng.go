package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

var inputFile string
var filename string
func main() {
	flag.StringVar(&inputFile, "input", "input.txt","The file with the input data.")
	flag.StringVar(&filename, "image", "image.png", "The path of the png file.")
	flag.Parse()
	fmt.Println(filename)
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	valid, header := ValidatePng(f)
	if valid {
		fout, err := os.Create("encrypted-" + filename)
		if err != nil {
			log.Fatal(err)
		}
		defer fout.Close()
		_, _ = fout.Write(header)
		chunk, err := ReadChunk(f)
		if err != nil {
			log.Fatal(err)
		}
		for chunk.name != "IDAT" {
			fmt.Printf("l: %d, n: %s, c: %d\n", chunk.length,  chunk.name, chunk.crc)
			_, _ = fout.Write(chunk.raw)
			chunk, err = ReadChunk(f)
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
		cryptChunk := CreateChunk(inputData, "crPt")
		_, _ = fout.Write(cryptChunk.raw)
		for {
			_, _ = fout.Write(chunk.raw)
			chunk, err = ReadChunk(f)
			if err != nil {
				break
			}
		}
	} else {
		log.Fatal("Invalid png.")
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

// encrypt and decrypt functions taken from
// https://stackoverflow.com/questions/18817336/golang-encrypting-a-string-with-aes-and-base64

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

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}