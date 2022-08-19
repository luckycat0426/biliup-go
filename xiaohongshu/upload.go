package xiaohongshu

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type xiaohongshu struct {
	Header http.Header
	Client http.Client
}

func UploadPhoto(f *os.File) {

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

func (x xiaohongshu) XiaoHongShuPreUploadQuery() (*CosEndPointInfo, error) {
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
	err = json.Unmarshal(body, t)
	if err != nil {
		return nil, err
	}
	if !t.Success {
		return nil, errors.New("cos preupload query failed")
	}
	return &t.Data.UploadTempPermits[0], nil
}
func GetCosAuthorization(c *CosEndPointInfo) string {
	t := fmt.Sprintf("%d;%d", time.Now().Unix(), c.ExpireTime/1000)
	a := CosAuthorization{
		QSignAlgorithm: "sha1",
		QAk:            c.SecretId,
		QSignTime:      t,
		QKeyTime:       t,
		QHeaderList:    "host",
		QUrlParamList:  "prefix;uploads",
		QSignature:     c.SecretKey,
	}
	return fmt.Sprintf("q-sign-algorithm=%s&q-ak=%s&q-sign-time=%s&q-key-time=%s&q-header-list=%s&q-url-param-list=%s&q-signature=%s", a.QSignAlgorithm, a.QAk, a.QSignTime, a.QKeyTime, a.QHeaderList, a.QUrlParamList, a.QSignature)
}

var DefaultHeader = http.Header{
	"User-Agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/63.0.3239.108"},
	"Referer":    []string{"https://www.bilibili.com"},
	"Connection": []string{"keep-alive"},
}

func cos(file *os.File, ret CosEndPointInfo, ChunkSize int, thread int) (*string,error) {
	client := &http.Client{}
	client.Timeout = 5 * time.Second
	uploadUrl := fmt.Sprintf("https://%s/%s", ret.UploadAddr, ret.FileIds[0])
	req, _ := http.NewRequest("POST", uploadUrl, nil)
	req.Header = DefaultHeader.Clone()
	req.Header.Set("Authorization", GetCosAuthorization(&ret))
	req.Header.Set("x-cos-security-token", ret.Token)
	res,err:=client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	resxml := cosPreUploadXmlRes{}
	err = xml.Unmarshal(body, &resxml)
	if err != nil {
		fmt.Println("marshal Cos Videos XMl error", string(body))
		//return nil, err
	}
	if resxml.UploadID == ""

}
