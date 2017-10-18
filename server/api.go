package main

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

type JResp struct {
	Hash  string   `json:"hash"`
	Files []string `json:"file"`
}

type FileEntry struct {
	Filename  string `json:"filename"`
	Content64 string `json:"file"`
}

func startServer(address string) {
	router := mux.NewRouter()
	router.HandleFunc("/api/adv/get/", GetAdv).Methods("GET")
	router.HandleFunc("/api/adv/set/", SetAdv).Methods("POST")
	log.Fatal(http.ListenAndServe(address, router))
}

func isAuthorized(r *http.Request) (auth bool, err error) {
	token := r.Header.Get("Token")
	client := &http.Client{}
	url := Config.AuthVer
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("Error while creating request for auth %s", err)
	}
	req.Header.Set("Token", token)
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("Error while making request for auth %s", err)
	}
	if resp.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}

// GetAdv отправляет клиенту json, в котором указан список файлов в папке static/images и их хеш
func GetAdv(w http.ResponseWriter, r *http.Request) {
	authorized, err := isAuthorized(r)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("can't authorize %s", err)
		return
	}
	if authorized == false {
		http.Error(w, "Authorization required", http.StatusUnauthorized)
		log.Printf("user wasn't authorized %s", err)
		return
	}

	picsDir, err := filepath.Abs(Config.ImagesRoot)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while getting absolute path %s", err)
		return
	}
	var resp JResp
	folders, err := ioutil.ReadDir(Config.ImagesRoot)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while reading pictures directory %s", err)
		return
	}
	lastFolder := filepath.Join(picsDir, folders[len(folders)-1].Name())
	pics, err := ioutil.ReadDir(lastFolder)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while reading pictures directory %s", err)
		return
	}
	for _, fileInfo := range pics {
		picPath := filepath.Join(lastFolder, fileInfo.Name())
		resp.Files = append(resp.Files, picPath)
	}
	resp.Hash, err = makeHash(pics, lastFolder)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SetAdv - получает gzip json, в котором указаны имена и тела файлов, закодированных в base64
// проводит раскодирование и записывает полученные файлы в static/images, удаляя старое содержимое
func SetAdv(w http.ResponseWriter, r *http.Request) {
	var entries []FileEntry

	authorized, err := isAuthorized(r)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("can't authorize %s", err)
		return
	}
	if authorized == false {
		http.Error(w, "Authorization required", http.StatusUnauthorized)
		log.Printf("user wasn't authorized %s", err)
		return
	}
	re, err := gzip.NewReader(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	defer re.Close()
	decoder := json.NewDecoder(re)
	if err = decoder.Decode(&entries); err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while handling directories %s", err)
		return
	}
	tempDirPath := filepath.Join(Config.ImagesRoot, "/temp/")
	if err = os.Mkdir(tempDirPath, 0744); err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while handling directories %s", err)
		return
	}
	for _, entry := range entries {
		path := filepath.Join(tempDirPath, entry.Filename)
		content, err := base64.StdEncoding.DecodeString(entry.Content64)
		if err != nil {
			http.Error(w, "Internal service problems", http.StatusInternalServerError)
			log.Printf("error while encoding to base64 in SetAdv %s", err)
			return
		}
		ioutil.WriteFile(path, content, 0744)
	}

	if err := handleFolders(&entries); err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("can't decode json file in SetAdv %s", err)
		return
	}
}

func handleFolders(entries *[]FileEntry) error {
	folderContent, err := ioutil.ReadDir(Config.ImagesRoot)
	if err != nil {
		return fmt.Errorf("can't read images directory %s", err)
	}
	if len(folderContent) >= Config.MaxFolderNum+1 {
		fmt.Println(len(folderContent))
		foldPath := filepath.Join(Config.ImagesRoot, folderContent[0].Name())
		os.RemoveAll(foldPath)
	}
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05.00")
	newDirPath := filepath.Join(Config.ImagesRoot, timestamp)
	tempDirPath := filepath.Join(Config.ImagesRoot, "/temp/")
	err = os.Rename(tempDirPath, newDirPath)
	if err != nil {
		return fmt.Errorf("can't read images directory %s", err)
	}
	return nil
}

func makeHash(files []os.FileInfo, lastFolder string) (string, error) {
	var allFiles []byte
	for _, file := range files {
		picPath := filepath.Join(lastFolder, file.Name())
		content, err := os.Open(picPath)
		if err != nil {
			log.Printf("can't open file %s for hashing", picPath)
			return "", err
		}
		fileBytes, err := ioutil.ReadAll(content)
		if err != nil {
			log.Printf("can't read file %s for hashing", picPath)
			return "", err
		}
		allFiles = append(allFiles, fileBytes...)
	}
	byteHash := md5.Sum(allFiles)
	hash := hex.EncodeToString(byteHash[:])
	return hash, nil
}
