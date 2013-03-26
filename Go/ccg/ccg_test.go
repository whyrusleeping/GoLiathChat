package ccg

import (
	"testing"
	"io"
	"bytes"
	"math/rand"
	"runtime"
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

func TestAsyncBufferPool(t *testing.T) {
	runtime.GOMAXPROCS(4)
	//Fails occasionally with GOMAXPROCS > 1
	bp := NewBufferPool(32)
	t.Log("Starting")
	rc := make(chan bool)
	pass := true
	for j := 0; j < 100; j++ {
		go func() {
			num := byte(j)
			for i := 0; i < 50; i++ {
				r := bp.GetBuffer(20 + rand.Intn(60))
				for k := 0; k < len(r); k++ {
					r[k] = num
				}
				for k := 0; k < len(r); k++ {
					if r[k] != num {
						pass = false
					}
				}

				bp.Free(r)
			}
			rc <- true
		}()
	}
	for j := 0; j < 100; j++ {
		<-rc
	}
	if !pass {
		t.Fail()
	}
}
func BenchmarkBufferPool(b *testing.B) {
	bp := NewBufferPool(32)
	for i := 0; i < b.N; i++ {
		a := bp.GetBuffer(20000 + rand.Intn(30000))
		//useBuffer(a)
		a[3] = 6
		bp.Free(a)
	}
}

func BenchmarkNonBufferPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a := make([]byte, 20000 + rand.Intn(30000))
		//useBuffer(a)
		a[3] = 6
	}
}
