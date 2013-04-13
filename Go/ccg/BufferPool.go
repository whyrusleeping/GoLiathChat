package ccg

var bufPool = NewBufferPool(32)

//A type to manage the recycling of byte arrays
//currently (and most likely, always) not threadsafed
type BufferPool struct {
	open [][]byte
	emptyOnFree bool
	placeHolder []byte
	freeChan    chan []byte
}

func NewBufferPool(size int) *BufferPool {
	bp := BufferPool{make([][]byte, size), false, make([]byte, 0),make(chan []byte)}
	//go bp.freeDaemon()
	return &bp
}

//NOT USED AS OF NOW
func (bp *BufferPool) freeDaemon() {
	for {
		buff := <-bp.freeChan
		for i := 0; i < len(bp.open); i++ {
			if bp.open[i] == nil {
				bp.open[i] = buff[:cap(buff)]
				break
			}
		}
	}
}

func (bp *BufferPool) Free(b []byte) {
	//add the buffer to the free pool
	for i := 0; i < len(bp.open); i++ {
		if bp.open[i] == nil {
			bp.open[i] = b[:cap(b)]
			break
		}
	}
}

func (bp *BufferPool) GetBuffer(size int) []byte{
	//find a buffer with at least the given size
	//and return it
	for i := 0; i < len(bp.open); i++ {
		if bp.open[i] != nil && len(bp.open[i]) >= size {
			r := bp.open[i][:size]
			bp.open[i] = nil
			return r
		}
	}
	return make([]byte, size)
}

//Ideas
//
//Find a way to make this better at asynchronous stuff..
