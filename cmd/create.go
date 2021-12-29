package cmd

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/btcsuite/btcutil/base58"
	"github.com/jon4hz/prvbin/pkg/privatebin"
	"github.com/spf13/cobra"
)

const fallbackEditor = "vim"

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new paste",
	RunE:  create,
}

var createFlags struct {
	text           string
	file           string
	password       string
	attachment     string
	expire         string
	sourcecode     bool
	markdown       bool
	burn           bool
	opendiscussion bool
	compress       bool
}

func init() {
	createCmd.Flags().StringVarP(&createFlags.text, "text", "t", "", "text to upload")
	createCmd.Flags().StringVarP(&createFlags.file, "file", "f", "", "file to upload")
	createCmd.Flags().StringVarP(&createFlags.password, "password", "p", "", "password to protect the paste")
	createCmd.Flags().StringVarP(&createFlags.expire, "expire", "e", "1month", "expire time")
	createCmd.Flags().BoolVarP(&createFlags.sourcecode, "sourcecode", "s", false, "use source code highlighting")
	createCmd.Flags().BoolVarP(&createFlags.markdown, "markdown", "m", false, "use markdown syntax highlighting")
	createCmd.Flags().BoolVarP(&createFlags.burn, "burn", "b", false, "burn after reading")
	createCmd.Flags().BoolVarP(&createFlags.opendiscussion, "opendiscussion", "d", false, "allow discussion for the paste")
	createCmd.Flags().StringVarP(&createFlags.attachment, "attachment", "a", "", "attachment file")
	createCmd.Flags().BoolVarP(&createFlags.compress, "compress", "c", true, "compress the file")
}

func create(cmd *cobra.Command, args []string) error {
	if !privatebin.ValidExpire(createFlags.expire) {
		return errors.New("invalid expire time")
	}

	var (
		content        []byte
		attachment     string
		attachmentName string
		burn           int
		opendiscussion int
		formatter      string
		err            error
	)

	if createFlags.text != "" {
		content = []byte(createFlags.text)
	} else if createFlags.file != "" {
		// read file
		content, err = ioutil.ReadFile(createFlags.file)
		if err != nil {
			return err
		}
	} else {
		content, err = createTmpFile()
		if err != nil {
			return err
		}
	}
	if len(content) == 0 {
		return errors.New("wont upload empty file")
	}

	if createFlags.burn {
		burn = 1
	}

	if createFlags.opendiscussion {
		opendiscussion = 1
	}

	switch {
	case createFlags.markdown:
		formatter = "markdown"
	case createFlags.sourcecode:
		formatter = "syntaxhighlighting"
	default:
		formatter = "plaintext"
	}

	if createFlags.attachment != "" {
		c, err := ioutil.ReadFile(createFlags.attachment) //nolint:govet
		if err != nil {
			return err
		}
		mimeType := http.DetectContentType(c)
		attachment = fmt.Sprintf("data:%s;base64,", mimeType)
		attachment += base64.StdEncoding.EncodeToString(c)

		attachmentName = filepath.Base(createFlags.attachment)
	}

	if rootFlags.url == "" {
		rootFlags.url = privatebin.DefaultURL
	}
	client := privatebin.NewClient(rootFlags.url)

	paste := privatebin.NewPaste(
		content, createFlags.password, formatter, attachmentName, attachment,
		createFlags.compress, burn, opendiscussion, createFlags.expire,
	)

	r, err := client.Send(paste)
	if err != nil {
		return err
	}

	fullURL := fmt.Sprintf("%s%s#%s", rootFlags.url, r.URL, base58.Encode(paste.GetPassphrase()))
	fmt.Println(fullURL)

	return nil
}

func createTmpFile() ([]byte, error) {
	// create a temporary file
	tmpF, err := ioutil.TempFile(".", ".prvbin-")
	if err != nil {
		return nil, err
	}
	tmpF.Close()

	// ensure the file gets deleted
	defer func() {
		os.Remove(tmpF.Name())
	}()

	// get the preferred editor or use vim as fallback
	editor, err := selectEditor()
	if err != nil {
		return nil, err
	}

	// open the temporary file in that editor
	edit := exec.Command(editor, tmpF.Name())
	edit.Stdin = os.Stdin
	edit.Stdout = os.Stdout
	edit.Stderr = os.Stderr
	if err = edit.Run(); err != nil {
		return nil, err
	}

	// read the file
	content, err := ioutil.ReadFile(tmpF.Name())
	if err != nil {
		return nil, err
	}
	return content, nil
}

func selectEditor() (string, error) {
	e := os.Getenv("EDITOR")
	if e == "" {
		e = fallbackEditor
	}
	return exec.LookPath(e)
}
