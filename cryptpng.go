package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"math"
	"strings"

	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/ssh/terminal"
	"github.com/cheggaaa/pb/v3"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const saltChunkName = "saLt"
const chunkName = "crPt"
const chunkSize = 0x100000
const scrN = 32768
const scrR = 8
const scrP = 1
const scrKeyLength = 32

var inputFile string
var outputFile string
var imageFile string
var decryptImage bool
func main() {
	encryptFlags := flag.NewFlagSet("encrypt", flag.ContinueOnError)
	decryptFlags := flag.NewFlagSet("decrypt", flag.ContinueOnError)
	encryptFlags.StringVar(&imageFile, "image", "image.png", "The path of the png file.")
	decryptFlags.StringVar(&imageFile, "image", "image.png", "The path of the png file.")
	encryptFlags.StringVar(&inputFile, "in", "input.txt","The file with the input data.")
	encryptFlags.StringVar(&outputFile, "out", "out.png", "The output filename for the image.")
	decryptFlags.StringVar(&outputFile, "out", "out.png", "The output file for the decrypted data.")
	flag.Parse()
	switch os.Args[1] {
	case "encrypt":
		err := encryptFlags.Parse(os.Args[2:])
		check(err)
		f, err := os.Open(imageFile)
		check(err)
		defer f.Close()
		check(err)
		if _, err := os.Stat(outputFile); err == nil {
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("The output file %s exists and will be overwritten. Continue? [Y/n] ", outputFile)
			if ans, _ := reader.ReadString('\n'); strings.ToLower(ans) != "y\n" {
				log.Fatal("Aborting...")
			}
		}
		fout, err := os.Create(outputFile)
		check(err)
		defer fout.Close()
		fin, err := os.Open(inputFile)
		check(err)
		defer fin.Close()
		EncryptDataPng(f, fin, fout)
	case "decrypt":
		err := decryptFlags.Parse(os.Args[2:])
		check(err)
		f, err := os.Open(imageFile)
		check(err)
		defer f.Close()
		check(err)
		fout, err := os.Create(outputFile)
		check(err)
		defer fout.Close()
		DecryptDataPng(f, fout)
	}
}

// encrypts the data of fin inside the png (f) and writes it to fout
func EncryptDataPng(f *os.File, fin *os.File, fout *os.File) {
	png := PngData{}
	log.Println("Reading image file...")
	err := png.Read(f)
	check(err)
	log.Println("Reading input file...")
	inputData, err := ioutil.ReadAll(fin)
	check(err)
	log.Println("Encrypting data...")
	inputData, salt := encryptData(inputData)
	check(err)
	log.Println("Creating salt chunk...")
	saltChunk := CreateChunk(salt, saltChunkName)
	png.AddMetaChunk(saltChunk)
	chunkCount := int(math.Ceil(float64(len(inputData)) / chunkSize))
	bar := pb.StartNew(chunkCount)
	log.Printf("Creating %d chunks to store the data...\n", chunkCount)
	for i := 0; i < chunkCount; i++ {
		dataStart := i * chunkSize
		dataEnd := dataStart + int(math.Min(chunkSize, float64(len(inputData[dataStart:]))))
		cryptChunk := CreateChunk(inputData[dataStart:dataEnd], chunkName)
		png.AddMetaChunk(cryptChunk)
		bar.Increment()
	}
	bar.Finish()
	log.Println("Writing output file...")
	err = png.Write(fout)
	log.Println("Finished!")
	check(err)
}

// Decrypts the data from a png file
func DecryptDataPng(f *os.File, fout *os.File) {
	png := PngData{}
	log.Println("Reading image file...")
	err := png.Read(f)
	check(err)
	salt := make([]byte, 0)
	log.Println("Getting salt chunk...")
	saltChunk := png.GetChunk(saltChunkName)
	if saltChunk != nil {
		salt = append(salt, saltChunk.data...)
	}
	var data []byte
	cryptChunks := png.GetChunksByName(chunkName)
	chunkCount := len(cryptChunks)
	log.Printf("Reading %d crypt chunks...", chunkCount)
	bar := pb.StartNew(chunkCount)
	for i, cryptChunk := range cryptChunks {
		if !cryptChunk.Verify() {
			log.Fatalf("Corrupted chunk data, chunk #%d", i)
		}
		data = append(data, cryptChunk.data...)
		bar.Increment()
	}
	bar.Finish()
	if len(data) > 0 {
		log.Println("Decrypting data...")
		data, err = decryptData(data, salt)
		check(err)
		log.Println("Writing output file...")
		_, err = fout.Write(data)
		check(err)
		log.Println("Finished!")
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
	if passwordSalt != nil {
		key, err := scrypt.Key(bytePw, *passwordSalt, scrN, scrR, scrP, scrKeyLength)
		check(err)
		return key, *passwordSalt
	} else {
		salt := make([]byte, 32)
		_, err = io.ReadFull(rand.Reader, salt)
		check(err)
		key, err := scrypt.Key(bytePw, salt, scrN, scrR, scrP, scrKeyLength)
		check(err)
		return key, salt
	}
}

// function to encrypt the data
func encrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cipherText := make([]byte, aes.BlockSize+len(data))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], data)
	return cipherText, nil
}

// function to decrypt the data
func decrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data) < aes.BlockSize {
		return nil, errors.New("data too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(data, data)
	return data, nil
}