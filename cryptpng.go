package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"math"

	"golang.org/x/crypto/ssh/terminal"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const saltChunkName = "saLt"
const chunkName = "crPt"
const chunkSize = 0x100000

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

// encrypts the data of fin inside the png (f) and writes it to fout
func EncryptDataPng(f *os.File, fin *os.File, fout *os.File) {
	png := PngData{}
	err := png.Read(f)
	check(err)
	inputData, err := ioutil.ReadAll(fin)
	check(err)
	inputData, salt := encryptData(inputData)
	check(err)
	saltChunk := CreateChunk(salt, saltChunkName)
	png.AddMetaChunk(saltChunk)
	chunkCount := int(math.Ceil(float64(len(inputData)) / chunkSize))
	for i := 0; i < chunkCount; i++ {
		dataStart := i * chunkSize
		dataEnd := dataStart + int(math.Min(chunkSize, float64(len(inputData[dataStart:]))))
		cryptChunk := CreateChunk(inputData[dataStart:dataEnd], chunkName)
		png.AddMetaChunk(cryptChunk)
	}
	err = png.Write(fout)
	check(err)
}

// Decrypts the data from a png file
func DecryptDataPng(f *os.File, fout *os.File) {
	png := PngData{}
	err := png.Read(f)
	check(err)
	salt := make([]byte, 0)
	saltChunk := png.GetChunk(saltChunkName)
	if saltChunk != nil {
		salt = append(salt, saltChunk.data...)
	}
	var data []byte
	for _, cryptChunk := range png.GetChunksByName(chunkName) {
		data = append(data, cryptChunk.data...)
	}
	if len(data) > 0 {
		data, err = decryptData(data, salt)
		if err != nil {
			log.Println("\nThe provided password is probably incorrect.")
		}
		check(err)
		_, err = fout.Write(data)
		check(err)
	} else {
		log.Fatal("no encrypted data inside the input image")
	}
}

// creates an encrypted png chunk
func encryptData(data []byte) ([]byte, []byte) {
	key, salt := readPassword(nil)
	encData, err := encrypt(key, data)
	check(err)
	return encData, salt
}

// decrypts the data of a png chunk
func decryptData(data []byte, salt []byte) ([]byte, error) {
	key, _ := readPassword(&salt)
	return decrypt(key, data)
}

// reads a password from the terminal
// turns off the input for the typing of the password
func readPassword(passwordSalt *[]byte) ([]byte, []byte) {
	fmt.Print("Password: ")
	bytePw, err := terminal.ReadPassword(int(syscall.Stdin))
	check(err)
	hash := sha512.New512_256()
	if passwordSalt != nil {
		hash.Write(append(*passwordSalt, bytePw...))
		return hash.Sum(nil), *passwordSalt
	} else {
		salt := make([]byte, 32)
		_, err = io.ReadFull(rand.Reader, salt)
		check(err)
		hash.Write(append(salt, bytePw...))
		return hash.Sum(nil), salt
	}
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