package privatebin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	DefaultURL = "https://bin.0xfc.de"
)

var validExpires = map[string]time.Duration{
	"5min":   5 * time.Minute,
	"10min":  10 * time.Minute,
	"1hour":  1 * time.Hour,
	"1day":   24 * time.Hour,
	"1week":  7 * 24 * time.Hour,
	"1month": 30 * 24 * time.Hour,
	"1year":  365 * 24 * time.Hour,
	"never":  0,
}

func ValidExpire(expire string) bool {
	_, ok := validExpires[expire]
	return ok
}

type Client struct {
	http *http.Client
	url  string
}

func NewClient(url string) *Client {
	return &Client{
		http: &http.Client{},
		url:  url,
	}
}

func NewPaste(content []byte, password string, formatter string, attachmentName string, attachment string,
	compress bool, burn int, opendiscussion int, expire string) *Paste {
	return &Paste{
		content:        content,
		password:       password,
		formatter:      formatter,
		attachmentName: attachmentName,
		attachment:     attachment,
		compress:       compress,
		burn:           burn,
		opendiscussion: opendiscussion,
		expire:         expire,
	}
}

func (c *Client) Send(p *Paste) (*RespSuccess, error) {
	adata, cyperText, err := p.encrypt()
	if err != nil {
		return nil, err
	}

	data := PasteReq{
		V:     2,
		AData: adata,
		CT:    base64.StdEncoding.EncodeToString(cyperText),
		Meta: PasteMeta{
			Expire: p.expire,
		},
	}

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Requested-With", "JSONHttpRequest")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respStatus RespStatus
	if err := json.Unmarshal(body, &respStatus); err != nil {
		return nil, err
	}

	if respStatus.Status != 0 {
		var respError RespError
		if err := json.Unmarshal(body, &respError); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%d: %s", respError.Status, respError.Message)
	}

	var respSuccess RespSuccess
	if err := json.Unmarshal(body, &respSuccess); err != nil {
		return nil, err
	}

	return &respSuccess, nil
}

func (p Paste) GetPassphrase() []byte {
	return p.passphrase
}
