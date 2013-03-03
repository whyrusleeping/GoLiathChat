package ccg

import (
	"testing"
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
