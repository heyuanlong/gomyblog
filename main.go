package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/russross/blackfriday.v2"
)

type Table map[string]interface{}
type Content struct {
	Text string
}

func main() {
	//input := []byte("Hello.\n\n* This is markdown.\n* It is fun\n* Love it or leave it.")
	//unsafe := blackfriday.Run(input)
	//html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	//fmt.Println(string(unsafe))
	//fmt.Println(string(html))
	//fmt.Println(input)
	GetFilelist("bin/source")
}

func GetFilelist(path string) {
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			DealDir(path, f)
		} else {
			DealFile(path, f)
		}
		fmt.Println(path)
		return nil
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
}
func DealDir(path string, f os.FileInfo) {
	distDirectory := strings.Replace(path, SOURCE_DIRECTORY, DIST_DIRECTORY, 1)
	exist, _ := PathExists(distDirectory)
	if !exist {
		os.Mkdir(distDirectory, os.ModePerm)
	}
}

func DealFile(path string, f os.FileInfo) {
	if f.ModTime().Unix() < (time.Now().Unix() - EXPIRE_TIME) {
		return
	}
	distFile := strings.Replace(path, SOURCE_DIRECTORY, DIST_DIRECTORY, 1)
	distFile = strings.Replace(distFile, SOURCE_SUFFIX, DIST_SUFFIX, 1)

	sourceF, err := os.Open(path)
	if err != nil {
		fmt.Println("OpenFile error: ", err)
		return
	}
	defer sourceF.Close()
	distF, err := os.Create(distFile)
	if err != nil {
		fmt.Println("OpenFile error: ", err)
		return
	}
	defer distF.Close()

	buf := make([]byte, 102400)
	for {
		n, err := sourceF.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Read error: ", err)
			return
		}
		if n == 0 {
			break
		}

		str := strings.Replace(string(buf[0:n]), "\r", "", -1)
		html := Transform([]byte(str))

		var data Table = make(Table)
		var c Content
		c.Text = string(html)
		data["Blog"] = c
		allHtml := TemplateHtml("bin/templates/default/article.html", data)
		//fmt.Println(string(html))
		//fmt.Println(buf[0:n])
		//fmt.Println([]byte(str))
		if _, err := distF.Write(allHtml); err != nil {
			fmt.Println("Write error: ", err)
			return
		}
	}
	return
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func WithOptions(flags blackfriday.HTMLFlags) blackfriday.Option {
	params := blackfriday.HTMLRendererParameters{Flags: flags}
	renderer := blackfriday.NewHTMLRenderer(params)
	return blackfriday.WithRenderer(renderer)
}
func Transform(source []byte) []byte {
	flags := blackfriday.CommonHTMLFlags | blackfriday.TOC
	unsafe := blackfriday.Run(source, WithOptions(flags))
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	return html
}

func TemplateHtml(templateFile string, data Table) []byte {
	tpl, err := template.ParseFiles(templateFile)
	if err != nil {
		fmt.Println("template error: ", err)
		return []byte{}
	}
	bytesBuffer := bytes.NewBuffer([]byte{})
	tpl.Execute(bytesBuffer, data)
	return bytesBuffer.Bytes()
}
