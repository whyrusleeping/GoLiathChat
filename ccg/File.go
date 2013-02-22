package ccg

import (
	"bytes"
	"os"
)

//A struct to represent a File broken into blocks for transfer
type File struct {
	Filename string
	blocks   int32
	data     []*block
}

//A block of file data tagged with its index
type block struct {
	blockNum uint32
	data     []byte
}

//const BlockSize = 32768
const BlockSize = 8

//Loads the given file from the hard drive and breaks into blocks
func LoadFile(path string) (*File, error) {
	//Open File
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	//Get file info and calculate block count
	finfo, _ := os.Stat(path)
	size := finfo.Size()
	numBlocks := size / BlockSize
	if size%BlockSize != 0 {
		numBlocks++
	}

	//Create the file object
	rf := File{finfo.Name(), 0, make([]*block, numBlocks)}

	//Read the file into blocks
	blockCount := 0
	for ; size >= BlockSize; blockCount++ {
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
	fi, err := os.Create(f.Filename)
	if err != nil {
		return err
	}
	for i := 0; i < len(f.data); i++ {
		fi.Write(f.data[i].data)
	}
	fi.Close()
	return nil
}

//checks to see if the file is complete
func (f *File) IsComplete() bool {
	for i := 0; i < len(f.data); i++ {
		if f.data[i] == nil {
			return false
		}
	}
	return true
}

//Gets metadata about the file for sending over the network
//Packs Filename and number of blocks into the returned array
//Also packs in a byte flag to signal compression
func (f *File) getInfo() []byte {
	buf := new(bytes.Buffer)
	buf.Write(BytesFromShortString(f.Filename))
	buf.Write(BytesFromInt32(int32(len(f.data))))
	buf.WriteByte(0)
	return buf.Bytes()
}

//Returns an array of bytes containing a chunk of the file
func (f *File) getBytesForBlock(num int) []byte {
	buf := new(bytes.Buffer)
	//Possibly replace this with a 'file id' integer negotiated with the server
	buf.Write(BytesFromShortString(f.Filename))
	buf.Write(BytesFromInt32(int32(num)))
	buf.Write(BytesFromInt32(int32(len(f.data[num].data))))
	buf.Write(f.data[num].data)
	return buf.Bytes()
}

//Some thoughts:
//
//when file size gets large enough we wont want to hold the entire file in memory
// To avoid this i think making the packets one at a time and sending them off would be ideal
// on the receiving side, the packets could be written to a file as they are received. 
//Im not entirely sure how this works for torrents, but i beleive its similar. (i certainly have downloaded files larger than my ram before)
