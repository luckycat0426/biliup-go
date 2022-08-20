package httpClient

import (
	"errors"
	"github.com/google/go-querystring/query"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"os"
	"strconv"
)

type ChunkInfo struct {
	Order int
	Etag  string
}

const ChunkUploadMaxRetryTimes = 5

type ChunkUploader struct {
	UploadId     string
	Chunks       int
	ChunkSize    int
	TotalSize    int
	Threads      int
	Url          string
	ChunkInfo    chan ChunkInfo
	UploadMethod string
	Header       http.Header
	File         *os.File
	MaxThread    chan struct{}
}
type chunkParams struct {
	UploadId   string `url:"uploadId"`
	Chunks     int    `url:"chunks"`
	Total      int    `url:"total"`
	Chunk      int    `url:"chunk"`
	Size       int    `url:"size"`
	PartNumber int    `url:"partNumber"`
	Start      int    `url:"start"`
	End        int    `url:"end"`
}

func (u *ChunkUploader) upload() error {
	group := new(errgroup.Group)
	for i := 0; i < u.Chunks; i++ {
		u.MaxThread <- struct{}{}
		buf := make([]byte, u.ChunkSize)
		bufSize, _ := u.File.Read(buf)
		chunk := chunkParams{
			UploadId:   u.UploadId,
			Chunks:     u.Chunks,
			Chunk:      i,
			Total:      u.TotalSize,
			Size:       bufSize,
			PartNumber: i + 1,
			Start:      i * u.ChunkSize,
			End:        i*u.ChunkSize + bufSize,
		}
		group.Go(func() error {
			return u.uploadChunk(buf, chunk)
		})
	}
	if err := group.Wait(); err != nil {
		close(u.ChunkInfo)
		return err
	}
	close(u.ChunkInfo)
	return nil
}
func (u *ChunkUploader) uploadChunk(data []byte, params chunkParams) error {
	defer func() {
		<-u.MaxThread
	}()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod("PUT")
	for k, v := range u.Header {
		req.Header.Set(k, v[0])
	}
	req.SetBodyRaw(data)
	vals, _ := query.Values(params)
	req.SetRequestURI(u.Url + "?" + vals.Encode())
	for i := 0; i <= ChunkUploadMaxRetryTimes; i++ {
		err := fasthttp.Do(req, resp)
		if err != nil || resp.StatusCode() != 200 {
			log.Println("上传分块出现问题，尝试重连")
			log.Println(err)
		} else {
			c := ChunkInfo{
				Order: params.PartNumber,
				Etag:  "",
			}
			if u.UploadMethod == "cos" {
				c.Etag = string(resp.Header.Peek("ETag"))
				//Upos不需要ETAG
			}
			u.ChunkInfo <- c
			//fasthttp.ReleaseResponse(resp)
			break
		}
		//fasthttp.ReleaseResponse(resp)
		if i == ChunkUploadMaxRetryTimes {
			log.Println("上传分块出现问题，重试次数超过限制")
			return errors.New(strconv.Itoa(u.Chunks) + "分块上传失败")
		}
	}

	return nil
}
