package privatebin

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

type RespStatus struct {
	Status int `json:"status"`
}

type RespError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type RespSuccess struct {
	Status      int    `json:"status"`
	ID          string `json:"id"`
	URL         string `json:"url"`
	DeleteToken string `json:"deletetoken"`
}
