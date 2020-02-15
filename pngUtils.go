package main

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
)


type ChunkData struct {
	length uint32
	name string
	data []byte
	crc uint32
}

func (c *ChunkData) GetRaw() []byte {
	var raw []byte
	lengthRaw := make([]byte, 4)
	crcRaw := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthRaw, c.length)
	binary.BigEndian.PutUint32(crcRaw, c.crc)
	nameRaw := []byte(c.name)
	raw = append(raw, lengthRaw...)
	raw = append(raw, nameRaw...)
	raw = append(raw, c.data...)
	raw = append(raw, crcRaw...)
	return raw
}

type PngData struct {
	header []byte
	chunks []ChunkData
}

// Reads the png data from a file into the struct
func (p *PngData) Read(f *os.File) error {
	valid, header := ValidatePng(f)
	if valid {
		p.header = header
		err := p.readChunks(f)
		if err != io.EOF {
			return err
		}
	} else {
		return errors.New("invalid png")
	}
	return nil
}

// writes all the data of the png into a new file
func (p *PngData) Write(f *os.File) error {
	_, err := f.Write(p.header)
	if err != nil {
		return err
	}
	err = p.writeChunks(f)
	return err
}

// reads all chunks from a png file.
// must be called after reading the header
func (p *PngData) readChunks(f *os.File) error {
	chunk, err := ReadChunk(f)
	for err == nil {
		p.chunks = append(p.chunks, chunk)
		chunk, err = ReadChunk(f)
	}
	p.chunks = append(p.chunks, chunk)
	return err
}

// writes all chunks to the given file
func (p *PngData) writeChunks(f *os.File) error {
	for _, chunk := range p.chunks {
		_, err := f.Write(chunk.GetRaw())
		if err != nil {
			return err
		}
	}
	return nil
}

// adds a meta chunk to the chunk data before the IDAT chunk.
func (p *PngData) AddMetaChunk(metaChunk ChunkData) {
	var newChunks []ChunkData
	appended := false
	for _, chunk := range p.chunks {
		if chunk.name == "IDAT" && !appended {
			newChunks = append(newChunks, metaChunk)
			newChunks = append(newChunks, chunk)
			appended = true
		} else {
			newChunks = append(newChunks, chunk)
		}
	}
	p.chunks = newChunks
}

// Returns the reference of a chunk by name
func (p *PngData) GetChunk(name string) *ChunkData {
	for _, chunk := range p.chunks {
		if chunk.name == name {
			return &chunk
		}
	}
	return nil
}

// returns all chunks with a given name
func (p *PngData) GetChunksByName(name string) []ChunkData {
	var chunks []ChunkData
	for _, chunk := range p.chunks {
		if chunk.name == name {
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}

// validates the png by reading the header of the file
func ValidatePng(f *os.File) (bool, []byte) {
	headerBytes := make([]byte, 8)
	_, err := f.Read(headerBytes)
	check(err)
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
	return ChunkData{
		length: length,
		name:   name,
		data:   data,
		crc:    crc,
	}, err
}

// creates a chunk with the given data and name
func CreateChunk(data []byte, name string) ChunkData {
	rawLength := make([]byte, 4)
	binary.BigEndian.PutUint32(rawLength, uint32(len(data)))
	rawName := []byte(name)
	var dataAndName []byte
	dataAndName = append(dataAndName, rawName...)
	dataAndName = append(dataAndName, data...)
	crc := crc32.ChecksumIEEE(dataAndName)
	rawCrc := make([]byte, 4)
	binary.BigEndian.PutUint32(rawCrc, crc)
	return ChunkData{
		length: uint32(len(data)),
		name:   name,
		data:   data,
		crc:    crc,
	}
}

