package ccg

import (
	"bytes"
	"code.google.com/p/go.crypto/scrypt"
	"crypto/rand"
	"encoding/binary"
	"io"
	"strings"
)

//Awesome salt thanks to travis lane.
var tSalt = "brownchocolatemoosecoffeelatte"

func ReadInt32(c io.Reader) int32 {
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

func ReadShortString(c io.Reader) (string, error) {
	l := make([]byte, 2)
	_, err := c.Read(l)
	if err != nil {
		return "", err
	}
	var r uint16
	buf := bytes.NewBuffer(l)
	binary.Read(buf, binary.LittleEndian, &r)
	str := make([]byte, r)
	c.Read(str)
	return string(str), nil
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
	pHash, err := scrypt.Key([]byte(password), []byte(tSalt), 16384, 19, 7, 32)
	if err != nil {
		panic(err)
	}
	return pHash
}

func GetUserBytesForAuthFile(u *User, pHash []byte) []byte {
	if len(pHash) != 32 {
		return nil
	}
	buf := new(bytes.Buffer)
	buf.Write([]byte("["))
	buf.Write([]byte(u.Username))
	buf.Write([]byte("]"))
	buf.Write(pHash)
	return buf.Bytes()
}

func extractCommand(pay string) string {
	i := strings.Index(pay, " ")
	return pay[1:i]
}
