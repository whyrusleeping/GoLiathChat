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

var bufPool4b = NewBufferPool(16)

func ReadInt32(c io.Reader) int32 {
	buf := bufPool4b.GetBuffer(4)
	c.Read(buf)
	r := int32(buf[0])
	r += int32(buf[1]) << 8
	r += int32(buf[2]) << 16
	r += int32(buf[3]) << 24
	bufPool4b.Free(buf)
	return r
}

func BytesToInt32(a []byte) int32 {
	n := 0
	n += int(a[0])
	n += int(a[1]) << 8
	n += int(a[2]) << 16
	n += int(a[3]) << 24
	return int32(n)
}


func WriteInt32(n int32) []byte {
	arr := make([]byte, 4)
	arr[0] = byte(n)
	arr[1] = byte(n >> 8)
	arr[2] = byte(n >> 16)
	arr[3] = byte(n >> 24)
	return arr
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
	strbuf := bufPool.GetBuffer(int(r))
	c.Read(strbuf)
	str := string(strbuf)
	bufPool.Free(strbuf)
	return str, nil
}

func ReadLongString(c io.Reader) ([]byte, error) {
	r := ReadInt32(c)
	str := make([]byte, r)
	c.Read(str)
	return str, nil
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
	pep := bufPool.GetBuffer(32)
	rand.Reader.Read(pep)
	return pep
}

func HashPassword(password string) []byte {
	pHash, err := scrypt.Key([]byte(password), []byte(tSalt), 16384, 9, 7, 32)
	if err != nil {
		panic(err)
	}
	return pHash
}

func extractCommand(pay string) string {
	i := strings.Index(pay, " ")
	if i < 0 {
		i = len(pay)
	}
	return pay[1:i]
}
