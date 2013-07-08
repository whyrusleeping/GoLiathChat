package goliath

type MessageLog struct {
	messages []*Packet
	count int
}

func NewLog(initialSize int) *MessageLog {
	l := MessageLog{make([]*Packet, initialSize), 0}
	return &l
}

//Add the given packet to the history list. Resize array if needed
func (l *MessageLog) PushMessage(p *Packet) {
	l.messages = append(l.messages, p)
	l.count++
}

func (l *MessageLog) AddEntryInOrder(p *Packet) {
	i := 0
	for ; i < l.count && l.messages[i].Timestamp < p.Timestamp; i++ {}
	for j := l.count; j > i; j-- {
		l.messages[j] = l.messages[j - 1]
	}
	l.messages[i] = p
}

func (l *MessageLog) LastNEntries(n int) []*Packet {
	if n > l.count {
		n = l.count
	}
	return l.messages[l.count - n:l.count]
}

func (l *MessageLog) Clear() {
	for i := 0; i < l.count; i++ {
		l.messages[i] = nil
	}
	l.count = 0
}
