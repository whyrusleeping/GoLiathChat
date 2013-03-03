package ccg

import ()

var bufPool = NewBufferPool(32)

type BufferPool struct {
	open [][]byte
	emptyOnFree bool
	placeHolder []byte
}

func NewBufferPool(size int) *BufferPool {
	bp := BufferPool{make([][]byte, size), false, make([]byte, 0)}
	return &bp
}

func (bp *BufferPool) Free(b []byte) {
	//add the buffer to the free pool
	for i := 0; i < len(bp.open); i++ {
		if bp.open[i] == nil {
			bp.open[i] = b[:cap(b)]
			return
		}
	}
}

/*func (bp *BufferPool) Free(b []byte) {
	//add the buffer to the free pool
	for i := 0; i < len(bp.open); i++ {
		if bp.open[i] == nil {
			if bp.emptyOnFree {
				bp.open[i] = bp.placeHolder
				go func() {
					plc := i
					b = b[:cap(b)]
					for j := 0; j < len(b); j++ {
						b[j] = 0
					}
					bp.open[plc] = b
				}()
			} else {
				bp.open[i] = b[:cap(b)]
				return
			}
		}
	}
}*/


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
