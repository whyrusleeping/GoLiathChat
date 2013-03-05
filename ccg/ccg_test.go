package ccg

import (
	"testing"
	"io"
	"bytes"
	"math/rand"
)

func TestPasswordHash(t *testing.T) {
	h1 := HashPassword("testPassword")
	h2 := HashPassword("testPassword")
	if len(h1) != len(h2) {
		t.Fail()
	}
	for i:= 0; i < len(h1); i++ {
		if h1[i] != h2[i] {
			t.Fail()
		}
	}
}

func TestInt32Serialize(t *testing.T) {
	num := int32(12345678)
	arr := WriteInt32(num)
	ver := BytesToInt32(arr)
	if num != ver {
		t.Fail()
	}
}

func TestInt32Stream(t *testing.T) {
	num := int32(6431527)
	buf := new(bytes.Buffer)
	buf.Write(WriteInt32(num))
	ver := ReadInt32(buf)
	if num != ver {
		t.Fail()
	}
}

func ReadInt32Alt(c io.Reader) int32 {
	buf := make([]byte,4)
	r := int32(buf[0])
	r += int32(buf[1]) << 8
	r += int32(buf[2]) << 16
	r += int32(buf[3]) << 24
	return r
}

func BenchmarkEntireIntParse(b *testing.B) {
	str := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		str.Write(WriteInt32(rand.Int31n(10e8)))
		ReadInt32Alt(str)
	}
}
func BenchmarkByteWiseIntParse(b *testing.B) {
	str := new(bytes.Buffer)
	for i := 0; i < b.N; i++ {
		str.Write(WriteInt32(rand.Int31n(10e8)))
		ReadInt32(str)
	}
}

func BenchmarkBufferPool(b *testing.B) {
	bp := NewBufferPool(32)
	for i := 0; i < b.N; i++ {
		a := bp.GetBuffer(2000 + rand.Intn(3000))
		//useBuffer(a)
		a[3] = 6
		bp.Free(a)
	}
}

func BenchmarkNonBufferPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := make([]byte, 2000 + rand.Intn(3000))
		//useBuffer(a)
		a[3] = 6
	}
}
