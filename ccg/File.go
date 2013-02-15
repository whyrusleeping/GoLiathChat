package ccg

import (
	"os"
	"bytes"
)

type File struct {
	filename string
	data []*block
}

type block struct {
	blockNum uint32
	data []byte
}

const BlockSize = 32768

//Loads the given file from the hard drive and breaks into blocks
func LoadFile(path string) (*File, error) {
	f,err := os.Open(path)
	if err != nil {
		return nil, err
	}
	finfo,_ := os.Stat(path)
	size := finfo.Size()
	numBlocks := size / BlockSize
	if size % BlockSize != 0 {
		numBlocks++
	}
	rf := File{}
	rf.filename = path
	rf.data = make([]*block, numBlocks)
	blockCount := 0
	for ;size >= BlockSize;blockCount++ {
		b := NewBlock(BlockSize)
		size -= BlockSize
		f.Read(b.data)
		b.blockNum = uint32(blockCount)
		rf.data[blockCount] = b
	}
	if size > 0 {
		b := NewBlock(int(size))
		f.Read(b.data)
		b.blockNum = uint32(blockCount)
		rf.data[blockCount] = b
	}
	return &rf, nil
}

//Creates a block with the given size (in bytes)
func NewBlock(size int) *block {
	b := block{}
	b.data = make([]byte, size)
	return &b
}

//Writes the file to the hard disk
func (f *File) Save() error {
	fi, err := os.Create(f.filename)
	if err != nil {
		return err
	}
	for i := 0; i < len(f.data); i++ {
		fi.Write(f.data[i].data)
	}
	fi.Close()
	return nil
}

//Gets metadata about the file for sending over the network
//Packs filename and number of blocks into the returned array
func (f *File) getInfo() []byte {
	buf := new(bytes.Buffer)
	buf.Write(BytesFromShortString(f.filename))
	buf.Write(BytesFromInt32(int32(len(f.data))))
	return buf.Bytes()
}

//Returns an array of bytes containing a chunk of the file
func (f *File) getBytesForBlock(num int) []byte {
	buf := new(bytes.Buffer)
	//Possibly replace this with a 'file id' integer negotiated with the server
	buf.Write(BytesFromShortString(f.filename))
	buf.Write(BytesFromInt32(int32(num)))
	buf.Write(BytesFromInt32(int32(len(f.data[num].data))))
	buf.Write(f.data[num].data)
	return buf.Bytes()
}
