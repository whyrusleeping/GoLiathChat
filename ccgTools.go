package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"code.google.com/p/go.crypto/scrypt"
	"net"
)

//Awesome salt thanks to travis lane.
var tSalt = "brownchocolatemoosecoffeelatte"

func ReadInt32(c net.Conn) int32 {
	var r int32
	buf := make([]byte, 4)
	c.Read(buf)
	obuf := bytes.NewBuffer(buf)
	binary.Read(obuf, binary.LittleEndian, &r)
	return r
}

func BytesFromInt32(i int32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, i)
	return buf.Bytes()
}

func BytesFromShortString(s string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint16(len(s)))
	for _, c := range s {
		binary.Write(buf, binary.LittleEndian, byte(c))
	}
	return buf.Bytes()
}

func BytesFromLongString(s string) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint32(len(s)))
	for _, c := range s {
		binary.Write(buf, binary.LittleEndian, byte(c))
	}
	return buf.Bytes()
}

func GeneratePepper() []byte {
	pep := make([]byte, 32)
	rand.Reader.Read(pep)
	return pep
}

func HashPassword(password string) []byte {
	pHash,_ := scrypt.Key([]byte(password), []byte(tSalt), 2^17, 19, 103, 32)
	return pHash
}
