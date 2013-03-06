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

func TestShortString(t *testing.T) {
	testStr := "abcdefghijklmnopqrstuvwxyztesting testing testing"
	buf := BytesFromShortString(testStr)
	stream := new(bytes.Buffer)
	stream.Write(buf)
	g,_ := ReadShortString(stream)
	if testStr != g {
		t.Fatal("Strings did not match")
	}
}

func TestPacketSerialization(t *testing.T) {
	p := NewPacket(TMessage, "TestName", []byte("text and things to what where why how"))
	buf := new(bytes.Buffer)
	buf.Write(p.GetBytes())
	np, err := ReadPacket(buf)
	if err != nil {
		t.Fatal(err)
	}
	check := true
	check = check && (p.Typ == np.Typ)
	check = check && (p.Timestamp == np.Timestamp)
	check = check && (p.Username == np.Username)
	check = check && (string(p.Payload) == string(np.Payload))
	if !check {
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
