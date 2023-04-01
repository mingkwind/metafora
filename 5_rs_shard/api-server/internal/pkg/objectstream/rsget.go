package objectstream

import (
	"fmt"
	"io"
	"sync"

	"github.com/klauspost/reedsolomon"
)

type decoder struct {
	readers   []io.Reader
	writers   []io.Writer
	enc       reedsolomon.Encoder
	size      int64 // 数据真实的大小
	cache     []byte
	cacheSize int // 当前cache的大小
	total     int64
}

func Newdecoder(reader []io.Reader, writer []io.Writer, size int64) *decoder {
	enc, _ := reedsolomon.New(DATA_SHARDS, PARITY_SHARDS)
	return &decoder{
		readers:   reader,
		writers:   writer,
		enc:       enc,
		size:      size,
		cache:     nil,
		cacheSize: 0,
		total:     0,
	}
}

func (d *decoder) getData() error {
	if d.total == d.size {
		// 说明已经读读取完啦
		return io.EOF
	}
	// 生成6个分片大小的shards
	shards := make([][]byte, ALL_SHARDS)
	repairIds := make([]int, 0)
	for i := range shards {
		if d.readers[i] == nil {
			repairIds = append(repairIds, i)
		} else {
			shards[i] = make([]byte, BLOCK_PER_SHARD)
			n, e := io.ReadFull(d.readers[i], shards[i])
			if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
				// 说明读取错误
				shards[i] = nil
			} else if n != BLOCK_PER_SHARD {
				// 说明读取不够，那就缩减大小
				shards[i] = shards[i][:n]
			}
		}
	}
	e := d.enc.Reconstruct(shards)
	if e != nil {
		return e
	}
	for _, i := range repairIds {
		// 对于丢失的分片，需要写入到writer中
		// 进行修复
		d.writers[i].Write(shards[i])
	}
	for i := 0; i < DATA_SHARDS; i++ {
		// 遍历4个数据分片，将每个分片的数据写入到cache中
		shardSize := int64(len(shards[i]))
		// 如果超出了size，那么就只取size的大小
		if d.total+shardSize > d.size {
			shardSize = d.size - d.total
		}
		d.cache = append(d.cache, shards[i][:shardSize]...)
		d.cacheSize += int(shardSize)
		d.total += shardSize
	}
	return nil
}

func (d *decoder) Read(p []byte) (n int, err error) {
	if d.cacheSize == 0 {
		// 说明cache为空，需要从reader中读取
		e := d.getData()
		if e != nil {
			// 说明无法获取更多的数据了
			return 0, e
		}
	}
	length := len(p)
	if d.cacheSize < length {
		length = d.cacheSize
	}
	d.cacheSize -= length
	// 将cache中的数据拷贝到p中
	copy(p, d.cache[:length])
	d.cache = d.cache[length:]
	return length, nil
}

type RSGetStream struct {
	*decoder
}

func NewRSGetStream(locateInfo map[int]string, dataServers []string, hash string, size int64) (*RSGetStream, error) {
	// locateInfo是获取到hash值的节点，dataserver是丢失分片的节点
	if len(locateInfo)+len(dataServers) != ALL_SHARDS {
		return nil, fmt.Errorf("dataServer number mismatch")
	}
	readers := make([]io.Reader, ALL_SHARDS)
	wg := &sync.WaitGroup{}
	wg.Add(ALL_SHARDS)
	var err error
	for i := 0; i < ALL_SHARDS; i++ {
		go func(i int) {
			defer wg.Done()
			server := locateInfo[i]
			if server == "" {
				// 说明数据为空，取上一个随机节点补上
				locateInfo[i] = dataServers[0]
				dataServers = dataServers[1:]
				return
			}
			reader, e := NewGetStream(server, fmt.Sprintf("%s.%d", hash, i))
			if e != nil && err == nil {
				err = e
			} else {
				readers[i] = reader
			}
		}(i)
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	writers := make([]io.Writer, ALL_SHARDS)
	perShard := (size + DATA_SHARDS - 1) / DATA_SHARDS
	// 因为他会进行填充的操作，所以这里需要进行一次修正
	for i := range readers {
		if readers[i] == nil {
			// 该节点为空，说明需要写入
			writers[i], err = NewTempPutStream(locateInfo[i], fmt.Sprintf("%s.%d", hash, i), perShard)
			if err != nil {
				return nil, err
			}
		}
	}

	dec := Newdecoder(readers, writers, size)
	return &RSGetStream{dec}, nil
}

func (s *RSGetStream) Close() {
	for i := range s.writers {
		if s.writers[i] != nil {
			s.writers[i].(*TempPutStream).Commit(true)
		}
	}
}
