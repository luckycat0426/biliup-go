package xiaohongshu

import (
	"biliup/httpClient"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	tCos "github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type xiaohongshu struct {
	Header    http.Header
	Client    *httpClient.Client
	ChunkSize int
	Threads   int
}

func New(cookies string) *xiaohongshu {
	x := xiaohongshu{Header: DefaultHeader,
		Client:    httpClient.New(nil),
		ChunkSize: 4 * 1024 * 1024,
		Threads:   4,
	}
	x.Header.Set("cookie", cookies)
	x.Client.Header = x.Header
	return &x
}
func (x *xiaohongshu) UploadPhoto(f *os.File) (*string, error) {
	c, err := x.XiaoHongShuPreUploadQuery()
	if err != nil {
		return nil, err
	}
	UploadFileName, err := cos(f, c, x.Client)
	if err != nil {
		return nil, err
	}
	return UploadFileName, nil
}

type cosPreUploadXmlRes struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	UploadID string   `xml:"UploadId"`
}
type CosEndPointInfo struct {
	FileIds    []string `json:"fileIds"`
	SecretId   string   `json:"secretId"`
	SecretKey  string   `json:"secretKey"`
	Token      string   `json:"token"`
	ExpireTime int      `json:"expireTime"`
	UploadAddr string   `json:"uploadAddr"`
}
type CosAuthorization struct {
	QSignAlgorithm string `json:"q-sign-algorithm"`
	QAk            string `json:"q-ak"`
	QSignTime      string `json:"q-sign-time"`
	QKeyTime       string `json:"q-key-time"`
	QHeaderList    string `json:"q-header-list"`
	QUrlParamList  string `json:"q-url-param-list"`
	QSignature     string `json:"q-signature"`
}

func (x *xiaohongshu) XiaoHongShuPreUploadQuery() (*CosEndPointInfo, error) {
	queryUrl := fmt.Sprintf("https://creator.xiaohongshu.com/api/media/v1/upload/web/permit?_=%f&biz_name=spectrum&scene=image&file_count=1", rand.Float64())
	req, _ := http.NewRequest("GET", queryUrl, nil)
	resp, err := x.Client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	t := struct {
		Success bool `json:"success"`
		Data    struct {
			UploadTempPermits []CosEndPointInfo `json:"uploadTempPermits"`
		} `json:"data"`
	}{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &t)
	if err != nil {
		return nil, err
	}
	if !t.Success {
		return nil, errors.New("cos preupload query failed")
	}
	return &t.Data.UploadTempPermits[0], nil
}

var DefaultHeader = http.Header{
	"User-Agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/63.0.3239.108"},
	"Referer":    []string{"https://www.xiaohongshu.com"},
	"Connection": []string{"keep-alive"},
}

func getPreUploadXml(req *http.Request, client *httpClient.Client) (*cosPreUploadXmlRes, error) {

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	resXml := cosPreUploadXmlRes{}
	err = xml.Unmarshal(body, &resXml)
	if err != nil {
		return nil, fmt.Errorf("unmarshal Cos Videos XMl error: %s", string(body))
	}
	if resXml.UploadID == "" {
		return nil, fmt.Errorf("failed to get uploadId: %s", string(body))
	}
	return &resXml, nil
}

func cos(file *os.File, ret *CosEndPointInfo, client *httpClient.Client) (*string, error) {
	uploadUrl := fmt.Sprintf("https://%s/%s", ret.UploadAddr, ret.FileIds[0])
	req, err := http.NewRequest("PUT", uploadUrl, file)
	if err != nil {
		return nil, err
	}
	duration := int64(ret.ExpireTime) - time.Now().Unix()
	tCos.AddAuthorizationHeader(ret.SecretId, ret.SecretKey, ret.Token, req, tCos.NewAuthTime(time.Duration(duration)))
	req.Header.Set("x-cos-security-token", ret.Token)
	resp, err := client.Do(req)
	body, err := io.ReadAll(resp.Body)
	fmt.Println(body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return &ret.FileIds[0], nil
}

func RequestCosMerge(req *http.Request, client *httpClient.Client) error {
	client.Timeout = 15 * time.Second
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("request Cos Merge Failed %s", body)
	}
	return nil
}

type partsXml struct {
	XMLName xml.Name `xml:"CompleteMultipartUpload"`
	Part    []struct {
		XMLName    xml.Name `xml:"Part"`
		PartNumber int      `xml:"PartNumber"`
		ETag       struct {
			Value string `xml:",innerxml"`
		} `xml:"ETag"`
	}
}
