package privatebin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/btcsuite/btcutil/base58"
)

const (
	defaultPasteURL = "https://paste.i2pd.xyz"
)

type PasteReq struct {
	V     int         `json:"v"`
	AData interface{} `json:"adata"`
	CT    string      `json:"ct"`
	Meta  PasteMeta   `json:"meta"`
}

type PasteMeta struct {
	Expire string `json:"expire"`
}

type RawPasteData struct {
	Paste          string `json:"paste"`
	Attachment     string `json:"attachment,omitempty"`
	AttachmentName string `json:"attachment_name,omitempty"`
}

type InnerPaste struct {
	Nonce           string
	KDFSalt         string
	KDFIterations   int
	KDFKeySize      int
	ADataSize       int
	CipherAlgo      string
	CipherMode      string
	CompressionType string
}

type PasteData struct {
	InnerPaste     InnerPaste
	Formatter      string
	OpenDiscussion int
	Burn           int
}

type Paste struct {
	url            string
	content        []byte
	passphrase     []byte
	password       string
	formatter      string
	attachmentName string
	attachment     string
	compress       bool
	burn           int
	opendiscussion int
	expire         string
}

type Opts func(*Paste)

func WithURL(url string) Opts {
	return func(p *Paste) {
		p.url = url
	}
}

func NewPaste(content []byte, password string, formatter string, attachmentName string, attachment string,
	compress bool, burn int, opendiscussion int, expire string, opts ...Opts) *Paste {
	p := &Paste{
		url:            defaultPasteURL,
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
	for _, f := range opts {
		f(p)
	}
	return p
}

type RespStatus struct {
	Status int `json:"status"`
}

type RespError struct {
	Status  int
	Message string
}

type RespSuccess struct {
	Status      int
	ID          string
	URL         string
	DeleteToken string
}

func (p *Paste) Send() error {
	adata, cyperText, err := p.encrypt()
	if err != nil {
		return err
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
		return err
	}
	fmt.Println(string(b))

	h := &http.Client{}
	req, err := http.NewRequest("POST", p.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("X-Requested-With", "JSONHttpRequest")

	resp, err := h.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("%d: %s", resp.StatusCode, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respStatus RespStatus
	if err := json.Unmarshal(body, &respStatus); err != nil {
		return err
	}

	if respStatus.Status != 0 {
		var respError RespError
		if err := json.Unmarshal(body, &respError); err != nil {
			return err
		}
		return fmt.Errorf("%d: %s", respError.Status, respError.Message)
	}

	var respSuccess RespSuccess
	if err := json.Unmarshal(body, &respSuccess); err != nil {
		return err
	}

	fmt.Printf("%s%s#%s\n", p.url, respSuccess.URL, base58.Encode(p.passphrase)) // TODO: return instead of print

	return nil
}
