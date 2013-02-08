package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"code.google.com/p/go.crypto/scrypt"
	"net"
)

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

func GeneratePepper() []byte {
	pep := make([]byte, 32)
	rand.Reader.Read(pep)
	return pep
}

func HashPassword(password string) []byte {
	salt := "brownchocolatemoosecoffeelatte"
	pHash := scrypt.Key([]byte(password), []byte(salt), 2^17, 19, 103, 32)
	return pHash
}
