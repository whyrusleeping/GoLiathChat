package ccg

import (
	"bytes"
	"os"
	"io"
	"io/ioutil"
	"compress/gzip"
)

//A struct to represent a File broken into blocks for transfer
type File struct {
	Filename string
	blocks   int32
	data     []*block
	compr byte
}

//A block of file data tagged with its index
type block struct {
	blockNum uint32
	data     []byte
}

//Having blocksize at 32768 causes a strange error i have yet to track down
//const BlockSize = 32768
const BlockSize = 4096

//Loads the given file from the hard drive and breaks into blocks
func LoadFile(path string) (*File, error) {
	//Open File
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	//Get file info and calculate block count
	finfo, _ := os.Stat(path)
	size := finfo.Size()
	compr := false
	if size > BlockSize {
		compr = true
	}
	var reader io.Reader
	if compr {
		//read in file and compress it
		arr,_ := ioutil.ReadAll(f)
		buff := new(bytes.Buffer)
		wr := gzip.NewWriter(buff)
		wr.Write(arr)
		wr.Close()
		reader = buff
		size = int64(buff.Len())
	} else {
		reader = f
	}

	//Calculate the number of blocks needed
	numBlocks := size / BlockSize
	if size%BlockSize != 0 {
		numBlocks++
	}

	//Create the file object
	cbyte := byte(0)
	if compr {
		cbyte = 1
	}
	rf := File{finfo.Name(), 0, make([]*block, numBlocks), cbyte}

	if compr {
		rf.Filename += ".gz"
	}

	//Read the file into blocks
	blockCount := 0
	for ; size >= BlockSize; blockCount++ {
		b := NewBlock(BlockSize)
		size -= BlockSize
		reader.Read(b.data)
		b.blockNum = uint32(blockCount)
		rf.data[blockCount] = b
	}
	if size > 0 {
		b := NewBlock(int(size))
		reader.Read(b.data)
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
	buf.Write(WriteInt32(int32(len(f.data))))
	buf.WriteByte(f.compr)
	return buf.Bytes()
}

//Returns an array of bytes containing a chunk of the file
func (f *File) getBytesForBlock(num int) []byte {
	buf := new(bytes.Buffer)
	//Possibly replace this with a 'file id' integer negotiated with the server
	buf.Write(BytesFromShortString(f.Filename))
	buf.Write(WriteInt32(int32(num)))
	buf.Write(WriteInt32(int32(len(f.data[num].data))))
	buf.Write(f.data[num].data)
	return buf.Bytes()
}

//File should contain an input/output file stream, and a gzip wrapper over that. 
//It should then have sendBlock, or receiveBlock and immediately write to the buffer to conserve ram usage
//Also: Always gzip. Who cares?
