package objectstream

import (
	"fmt"
	"io"

	"github.com/klauspost/reedsolomon"
)

const (
	DATA_SHARDS     = 4
	PARITY_SHARDS   = 2
	ALL_SHARDS      = DATA_SHARDS + PARITY_SHARDS
	BLOCK_PER_SHARD = 8000
	BLOCK_SIZE      = BLOCK_PER_SHARD * DATA_SHARDS
)

type encoder struct {
	writers []io.Writer
	enc     reedsolomon.Encoder
	cache   []byte
}

func NewEncoder(writer []io.Writer) *encoder {
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	return &encoder{
		writers: writer,
		enc:     enc,
		cache:   nil,
	}

}

func (e *encoder) Write(p []byte) (n int, err error) {
	length := len(p)
	current := 0
	for length != 0 {
		// next是cache的剩余空间
		next := BLOCK_SIZE - len(e.cache)
		if next > length {
			next = length
		}
		e.cache = append(e.cache, p[current:current+next]...)
		if len(e.cache) == BLOCK_SIZE {
			e.Flush()
		}
		current += next
		length -= next
	}
	return len(p), nil
}

func (e *encoder) Flush() {
	if len(e.cache) == 0 {
		return
	}
	shards, _ := e.enc.Split(e.cache)
	e.enc.Encode(shards)
	for i := range shards {
		e.writers[i].Write(shards[i])
	}
	e.cache = []byte{}
}

type RSPutStream struct {
	*encoder
}

func NewRSPutStream(dataServers []string, hash string, size int64) (*RSPutStream, error) {
	if len(dataServers) != ALL_SHARDS {
		return nil, fmt.Errorf("dataServers count %d != %d", len(dataServers), ALL_SHARDS)
	}
	perShard := (size + DATA_SHARDS - 1) / DATA_SHARDS // 进行向上取整
	writers := make([]io.Writer, ALL_SHARDS)
	var err error
	for i := 0; i < ALL_SHARDS; i++ {
		writers[i], err = NewTempPutStream(dataServers[i], fmt.Sprintf("%s.%d", hash, i), perShard)
		if err != nil {
			return nil, err
		}
	}
	enc := NewEncoder(writers)
	return &RSPutStream{enc}, nil
}

func (s *RSPutStream) Commit(success bool) {
	// 将剩余的数据上传
	s.Flush()
	for i := range s.writers {
		s.writers[i].(*TempPutStream).Commit(success)
	}
}
