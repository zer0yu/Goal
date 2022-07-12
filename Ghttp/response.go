package Ghttp

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// Execute 发送请求
func (h *Http) Execute() *http.Response {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	//fmt.Printf("HttpTransport: %p \n", h.HttpTransport)
	var err error
	// h.HttpClient.Transport = h.HttpTransport
	h.HttpResponse, err = h.HttpClient.Do(h.HttpRequest)
	if err != nil {
		log.Println("[!] Http Execute Error : ", err)
		h.HttpResponse = nil
		return nil
	}
	return h.HttpResponse
}

// Close 关闭请求与body
func (h *Http) Close() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if h.HttpResponse != nil {
		//log.Println("CLose(): ", h.HttpResponse.Body)
		//h.readAll()
		h.readNull()
		h.HttpResponse = nil
	}

	if h.CtxCancel != nil {
		h.CtxCancel()
	}

}

// GetRespHead 获取返回头
func (h *Http) GetRespHead(key string) string {
	if h.HttpResponse != nil {
		return h.HttpResponse.Header.Get(key)
	}
	return ""
}
func (h *Http) readNull() ([]byte, error) {
	_, err := io.Copy(io.Discard, h.HttpResponse.Body)
	return nil, err
}
func (h *Http) readAll() ([]byte, error) {
	//获取一个新的，如果不存在则会调用new创建
	buffer := pool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer func() {
		if buffer != nil {
			//重新放回去
			pool.Put(buffer)
			buffer = nil
		}
	}()
	if h.HttpResponse == nil {
		return nil, fmt.Errorf("HttpResponse is nil")
	}
	if h.HttpResponse.Body == nil {
		return nil, fmt.Errorf("HttpResponse.Body is nil")
	}
	_, err := io.Copy(buffer, h.HttpResponse.Body)
	if err != nil && err != io.EOF {
		//log.Printf("readAll io.copy failure error:%v \n", err)
		return nil, fmt.Errorf("readAll io.copy failure error:%v", err)
	}
	defer h.HttpResponse.Body.Close()
	// 如果是gzip压缩则解压
	if h.HttpResponse.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(h.HttpResponse.Body)
		if err != nil {
			log.Println("gzipDecode error: ", err)
			return buffer.Bytes(), nil
			// return nil, err
		}
		defer reader.Close()
		return ioutil.ReadAll(reader)
	}
	return buffer.Bytes(), nil
}

// string的返回值
func (h *Http) Text() (string, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in Text : ", r)
		}
	}()
	r, err := h.readAll()
	if err != nil {
		return "", err
	}
	return string(r), err
}

// byte的返回值
func (h *Http) Byte() ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in Byte : ", r)
		}
	}()
	return h.readAll()
}
func (h *Http) SaveToFile(file string) (bool, error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in SaveToFile : ", r)
		}
	}()
	if h.HttpResponse == nil {
		log.Println("[!] HttpResponse Is Closed")
		return false, fmt.Errorf("HttpResponse is nil")
	}
	var f *os.File
	f, err = os.Create(file)
	if h.HttpResponse.Body == nil {
		return false, fmt.Errorf("HttpResponse.Body is nil")
	}
	defer h.HttpResponse.Body.Close()
	if err == nil {
		defer f.Close()
		_, err = io.Copy(f, h.HttpResponse.Body)
	}
	if err == nil {
		return true, err
	}
	return false, err

}

// statuscode
func (h Http) StatusCode() int {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if h.HttpResponse != nil {
		return h.HttpResponse.StatusCode
	}
	return -1
}
func (h *Http) RespCookie() string {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in RespCookie : ", r)
		}
	}()
	if h.HttpResponse == nil || h.HttpResponse.Header == nil {
		return ""
	}
	cookies := h.HttpResponse.Header.Values("set-cookie")
	cks := strings.Join(cookies, "; ")
	return cks
}
