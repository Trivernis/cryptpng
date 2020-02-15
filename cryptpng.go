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
	"io/ioutil"
	"log"
	"os"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const chunkName = "crPt"

var inputFile string
var outputFile string
var imageFile string
var decryptImage bool
func main() {
	flag.StringVar(&imageFile, "image", "image.png", "The path of the png file.")
	flag.BoolVar(&decryptImage, "decrypt", false, "If the input image should be decrypted.")
	flag.StringVar(&outputFile, "out", "out.png", "The output file for the encrypted/decrypted data.")
	flag.StringVar(&inputFile, "in", "input.txt","The file with the input data.")
	flag.Parse()
	if decryptImage {
		f, err := os.Open(imageFile)
		check(err)
		defer f.Close()
		check(err)
		fout, err := os.Create(outputFile)
		check(err)
		defer fout.Close()
		DecryptDataPng(f, fout)
	} else {
		f, err := os.Open(imageFile)
		check(err)
		defer f.Close()
		check(err)
		fout, err := os.Create(outputFile)
		check(err)
		defer fout.Close()
		fin, err := os.Open(inputFile)
		check(err)
		defer fin.Close()
		EncryptDataPng(f, fin, fout)
	}
}

func EncryptDataPng(f *os.File, fin *os.File, fout *os.File) {
	png := PngData{}
	err := png.Read(f)
	check(err)
	inputData, err := ioutil.ReadAll(fin)
	check(err)
	inputData, err = encryptData(inputData)
	check(err)
	cryptChunk := CreateChunk(inputData, chunkName)
	png.AddMetaChunk(cryptChunk)
	err = png.Write(fout)
	check(err)
}

// Decrypts the data from a png file
func DecryptDataPng(f *os.File, fout *os.File) {
	png := PngData{}
	err := png.Read(f)
	check(err)
	cryptChunk := png.GetChunk(chunkName)
	if cryptChunk != nil {
		data, err := decryptData(cryptChunk.data)
		check(err)
		_, err = fout.Write(data)
		check(err)
	} else {
		log.Fatal("no encrypted data inside the input image")
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

func decryptData(data []byte) ([]byte, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Password: ")
	pw, _ := reader.ReadString('\n')
	key := make([]byte, 32 - len(pw))
	key = append(key, []byte(pw)...)
	return decrypt(key, data)
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