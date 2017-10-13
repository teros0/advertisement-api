package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type JResp struct {
	Hash  string   `json:"hash"`
	Files []string `json:"file"`
}

type FileEntry struct {
	Filename  string `json:"filename"`
	Content64 string `json:"file"`
}

const Token = "1token"

func main() {
	funcFlag := flag.String("function", "set", "function to call")
	urlFlag := flag.String("url", "http://localhost:8080/api/adv/set/", "url to call")
	filesFlag := flag.String("files", "[]", "files to send")
	flag.Parse()
	switch *funcFlag {
	case "set":
		doSet(*urlFlag, *filesFlag)
	case "get":
		doGet(*urlFlag)
	}
}

func doSet(url, files string) {
	var entries []FileEntry
	var fileNames []string
	if err := json.Unmarshal([]byte(files), &fileNames); err != nil {
		fmt.Println(err)
	}

	picsDir, err := filepath.Abs("./static/imagesSend")
	if err != nil {
		log.Println(err)
		return
	}
	for _, fileName := range fileNames {
		picPath := filepath.Join(picsDir, fileName)
		file, err := os.Open(picPath)
		if err != nil {
			log.Println(err)
			return
		}
		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Println(err)
			return
		}
		var entry FileEntry
		entry.Filename = fileName
		entry.Content64 = base64.StdEncoding.EncodeToString(content)
		entries = append(entries, entry)
	}
	respBytes, err := json.Marshal(entries)
	if err != nil {
		log.Println(err)
		return
	}
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err = gzipWriter.Write(respBytes); err != nil {
		log.Println(err)
		return
	}
	if err = gzipWriter.Close(); err != nil {
		log.Println(err)
		return
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Token", Token)
	res, err := client.Do(req)
	fmt.Println(res.Status, "SENT REQ")
}

func doGet(url string) {
	var JSONResp JResp
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Token", Token)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&JSONResp); err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%+v\n", JSONResp)
}
