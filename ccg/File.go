package ccg

import (
	"os"
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
	finfo := Stat(f)
	size := finfo.Size()
	numBlocks := size / BlockSize
	if size % BlockSize != 0 {
		numBlocks++
	}
	rf := File{}
	rf.filename = path
	rf.data = make([]*block, numBlocks)
	blockCount := 0
	for size >= BlockSize {
		b := NewBlock(BlockSize)
		size -= BlockSize
		f.Read(b.data)
		b.blockNum = blockCount
		rf.data[blockCount] = b
	}
	if size > 0 {
		b := NewBlock(size)
		f.Read(b.data)
		b.blockNum = blockCount
		rf.data = b
	}
	return &rf
}

func NewBlock(size int) *block {
	b := block{}
	b.data = make([]byte, size)
	return &b
}

//Writes the file to the hard disk
func (f *File) Save() error {
	f, err := os.Create(f.filename)
	if err != nil {
		return err
	}
	for i := 0; i < len(f.data); i++ {
		f.Write(f.data[i].data)
	}
	f.Close()
}

//Gets metadata about the file for sending over the network
//Packs filename and number of blocks into the returned array
func (f *File) getInfo() []byte {
	buf := new(bytes.Buffer)
	buf.Write(BytesFromShortString(f.filename)
	buf.Write(BytesFromInt32(len(f.data)))
	return buf.Bytes
}
