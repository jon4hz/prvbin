package privatebin

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

const (
	kdfIterations = 100000
	kdfSaltLength = 8
	kdfKeySize    = 256
	adataSize     = 128
	cipherAlgo    = "aes"
	cipherMode    = "gcm"
)

func getRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	rand.Seed(time.Now().UnixNano())
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (p *Paste) appendRandPassphrase() error {
	var err error
	p.passphrase, err = getRandomBytes(32)
	return err
}

func (p *Paste) encrypt() (interface{}, []byte, error) {
	if err := p.appendRandPassphrase(); err != nil {
		return nil, nil, err
	}

	var key []byte
	if p.password != "" {
		key = append(p.passphrase, []byte(p.password)...)
	} else {
		key = p.passphrase
	}

	kdfSalt, err := getRandomBytes(kdfSaltLength)
	if err != nil {
		return nil, nil, err
	}

	kdfKey := pbkdf2.Key(key, kdfSalt, kdfIterations, 32, sha256.New)

	data := RawPasteData{
		Paste:          string(p.content),
		Attachment:     p.attachment,
		AttachmentName: p.attachmentName,
	}

	rawPasteBlob, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}

	compressionType := "none"
	var pasteBlob []byte

	if p.compress {
		compressionType = "zlib"
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		if _, err = w.Write(rawPasteBlob); err != nil {
			return nil, nil, err
		}
		if err = w.Close(); err != nil {
			return nil, nil, err
		}
		pasteBlob = b.Bytes()
	} else {
		pasteBlob = rawPasteBlob
	}

	nonce, err := getRandomBytes(16)
	if err != nil {
		return nil, nil, err
	}

	pasteData := []PasteData{
		{
			InnerPaste: []InnerPaste{
				{
					Nonce:           base64.StdEncoding.EncodeToString(nonce),
					KDFSalt:         base64.StdEncoding.EncodeToString(kdfSalt),
					KDFIterations:   kdfIterations,
					KDFKeySize:      kdfKeySize,
					ADataSize:       adataSize,
					CipherAlgo:      cipherAlgo,
					CipherMode:      cipherMode,
					CompressionType: compressionType,
				},
			},
			Formatter:      p.formatter,
			OpenDiscussion: p.opendiscussion,
			Burn:           p.burn,
		},
	}

	i := toJSONArray(pasteData)
	pasteAData, err := json.Marshal(i)
	if err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(kdfKey)
	if err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, nil, err
	}

	cipherText := aesgcm.Seal(nil, nonce, pasteBlob, pasteAData)

	return i, cipherText, nil
}

func toJSONArray(d []PasteData) []interface{} {
	ii := []interface{}{
		d[0].InnerPaste[0].Nonce,
		d[0].InnerPaste[0].KDFSalt,
		d[0].InnerPaste[0].KDFIterations,
		d[0].InnerPaste[0].KDFKeySize,
		d[0].InnerPaste[0].ADataSize,
		d[0].InnerPaste[0].CipherAlgo,
		d[0].InnerPaste[0].CipherMode,
		d[0].InnerPaste[0].CompressionType,
	}
	return []interface{}{ii, d[0].Formatter, d[0].OpenDiscussion, d[0].Burn}
}
