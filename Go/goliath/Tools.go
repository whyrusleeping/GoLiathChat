package goliath

import (
	"bytes"
	"os"
	"code.google.com/p/go.crypto/scrypt"
	"crypto/rand"
	"encoding/binary"
	"io"
	"strings"
	"crypto/x509"
	"crypto/rsa"
	"encoding/pem"
	"time"
	"math/big"
	"crypto/x509/pkix"
	"errors"
)

//Awesome salt thanks to travis lane.
var tSalt = "brownchocolatemoosecoffeelatte"

func ReadInt32(c io.Reader) int32 {
	buf := make([]byte, 4)
	c.Read(buf)
	r := int32(buf[0])
	r += int32(buf[1]) << 8
	r += int32(buf[2]) << 16
	r += int32(buf[3]) << 24
	return r
}

func BytesToInt32(a []byte) int32 {
	var n int
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

func WriteShortString(w io.Writer, s string) {
	binary.Write(w, binary.LittleEndian, uint16(len(s)))
	w.Write([]byte(s))
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
	if r < 0 {
		return "", errors.New("Cannot have length < 0")
	}
	strbuf := make([]byte, r)
	c.Read(strbuf)
	str := string(strbuf)
	return str, nil
}

func ReadLongString(c io.Reader) ([]byte, error) {
	r := ReadInt32(c)
	if r < 0 {
		return nil, errors.New("length < 0")
	}
	str := make([]byte, r)
	c.Read(str)
	return str, nil
}

func BytesFromLongString(s string) []byte {
	buf := new(bytes.Buffer)
	WriteLongString(buf, []byte(s))
	return buf.Bytes()
}

func WriteLongString(w io.Writer, s []byte) {
	w.Write(WriteInt32(int32(len(s))))
	w.Write(s)
}

func GeneratePepper() []byte {
	pep := make([]byte, 32)
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

func TryLoadCert(filename, host string) (*x509.Certificate, error) {
	return nil, nil
}

func SaveCert(c *x509.Certificate) error {
	return nil
}

func MakeCert(host string)  (error) {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return err
	}

	now := time.Now()

	bin := GetBinDir()

	template := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"GoliathChat"},
		},
		NotBefore: now.Add(-5 * time.Minute).UTC(),
		NotAfter:  now.AddDate(1, 0, 0).UTC(), // valid for 1 year.

		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(bin + "cert.pem")
	if err != nil {
		return err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(bin + "key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
	return nil
}

func GetBinDir() string {
	binf := os.Args[0]
	lslsh := strings.LastIndex(binf, "/")
	return binf[:lslsh+1]
}
