package objectstream

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type PutStream struct {
	writer *io.PipeWriter
	c      chan error
}

type TempPutStream struct {
	Server string
	Uuid   string
}

func NewTempPutStream(server, hash string, size int64) (*TempPutStream, error) {
	// apiserver向dataserver发起请求，告诉dataserver要上传的文件的hash和大小
	// dataserver返回一个uuid，用于后续的上传
	request, err := http.NewRequest("POST", "http://"+server+"/temp/"+hash, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Size", fmt.Sprintf("%d", size))
	client := http.Client{}
	r, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dataServer return http code %d", r.StatusCode)
	}
	uuid, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return &TempPutStream{server, string(uuid)}, nil
}

func (w *TempPutStream) Write(p []byte) (n int, e error) {
	// 根据uuid，用patch的方式向dataserver上传文件，文件内容为p
	request, err := http.NewRequest("PATCH", "http://"+w.Server+"/temp/"+w.Uuid, strings.NewReader(string(p)))
	if err != nil {
		return 0, err
	}
	client := http.Client{}
	r, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	if r.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("dataServer return http code %d", r.StatusCode)
	}
	return len(p), nil
}

func (w *TempPutStream) Commit(good bool) {
	var method string
	switch good {
	case true:
		method = "PUT"
	case false:
		method = "DELETE"
	}
	request, _ := http.NewRequest(method, "http://"+w.Server+"/temp/"+w.Uuid, nil)
	client := http.Client{}
	client.Do(request)
}

// 流式写入的方法
// 参考https://studygolang.com/articles/29059
func NewPutStream(server, hash string) *PutStream {
	reader, writer := io.Pipe()
	c := make(chan error)
	go func() {
		request, _ := http.NewRequest("PUT", "http://"+server+"/objects/", reader)
		// 在header中加入hash
		request.Header.Set("Digest", "SHA-256="+hash)
		client := http.Client{}
		r, e := client.Do(request)
		if e == nil && r.StatusCode != http.StatusOK {
			e = fmt.Errorf("dataServer return http code %d", r.StatusCode)
		}
		c <- e
	}()
	return &PutStream{writer, c}
}

func (w *PutStream) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

// Close方法会在写入完成后，关闭writer，然后等待c的返回值
func (w *PutStream) Close() error {
	w.writer.Close()
	return <-w.c
}
