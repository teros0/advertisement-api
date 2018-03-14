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

const Token = "PY5r1ouOZiqRUldWj0W7v"

func main() {
	funcFlag := flag.String("function", "set", "function to call")
	urlFlag := flag.String("url", "", "url to call")
	filesFlag := flag.String("files", "[]", "files to send")
	flag.Parse()
	switch *funcFlag {
	case "set":
		if *urlFlag == "" {
			*urlFlag = "http://localhost:8080/api/set-adv"
		}
		doSet(*urlFlag, *filesFlag)
	case "get":
		if *urlFlag == "" {
			*urlFlag = "http://localhost:8080/api/get-adv"
		}
		doGet(*urlFlag)
	default:
		fmt.Println("Choose an existing function, get or set")
		log.Fatalln("User tried to use uknown funciton")
	}
}

// Отсылает на /api/adv/set контент файлов, указанных в флаге -files.
// Файлы должны хранится в папке static/imagesSend, их имена и содержимое (в base64)
// отсылаются как json, который архивируется с помощью gzip
func doSet(url, files string) {
	var entries []FileEntry
	var fileNames []string
	fmt.Printf("TYPE: %T, VALUE: %+v\n", files, files)
	if err := json.Unmarshal([]byte(files), &fileNames); err != nil {
		log.Fatalf("Error while unmarshalling files argument %s", err)
	}
	if len(fileNames) == 0 {
		fmt.Println("Provide at least one filename")
		log.Fatalln("User tried to set zero pictures")
	}

	picsDir, err := filepath.Abs("images/")
	if err != nil {
		log.Fatalf("Error while getting abs path to pictures %s", err)
	}
	for _, fileName := range fileNames {
		picPath := filepath.Join(picsDir, fileName)
		file, err := os.Open(picPath)
		if err != nil {
			log.Fatalf("Error while opening file %s: %s", picPath, err)
		}
		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatalf("Can't read file %s: %s", picPath, err)
		}
		var entry FileEntry
		entry.Filename = fileName
		entry.Content64 = base64.StdEncoding.EncodeToString(content)
		entries = append(entries, entry)
	}
	respBytes, err := json.Marshal(entries)
	if err != nil {
		log.Fatalf("Error while marshalling files %s", err)
	}
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err = gzipWriter.Write(respBytes); err != nil {
		log.Fatalf("Error while trying to write to gzipWriter %s", err)
	}
	if err = gzipWriter.Close(); err != nil {
		log.Fatalf("Error while closing gzipWriter %s", err)
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		log.Fatalf("Error while creating request %s", err)
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Token", Token)
	_, err = client.Do(req)
	if err != nil {
		log.Fatalf("Error while executing request %s", err)
	}
}

// doGet - делает запрос к api/adv/get, получает список картинок, которые хранятся на сервере и их хеш
func doGet(url string) {
	var JSONResp JResp
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error while creating request %s", err)
	}
	req.Header.Set("Token", Token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error while executing request %s", err)
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(&JSONResp); err != nil {
		log.Fatalf("Error while decoding response %s", err)
	}
	fmt.Printf("%+v\n", JSONResp)
}
