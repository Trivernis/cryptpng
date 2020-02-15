package main

import (
	"encoding/binary"
	"hash/crc32"
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


// validates the png by reading the header of the file
func ValidatePng(f *os.File) (bool, []byte) {
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
func ReadChunk(f *os.File) (ChunkData, error) {
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
func CreateChunk(data []byte, name string) ChunkData {
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
		raw:    fullData,
	}
}

