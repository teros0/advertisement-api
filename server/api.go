package main

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	router.HandleFunc("/api/adv/get/", authHandler(GetAdv)).Methods("GET")
	router.HandleFunc("/api/adv/set/", authHandler(SetAdv)).Methods("POST")
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

func authHandler(f func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		f(w, r)
	}
}

// GetAdv отправляет клиенту json, в котором указан список файлов в папке static/images и их хеш
func GetAdv(w http.ResponseWriter, r *http.Request) {
	var resp JResp
	var lastFolderPath string
	picsDir, err := filepath.Abs(Config.ImagesRoot)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while getting absolute path %s", err)
		return
	}
	folders, err := ioutil.ReadDir(Config.ImagesRoot)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while reading pictures directory %s", err)
		return
	}
	lastFolder := folders[len(folders)-1]
	if lastFolder.Name() == "temp" {
		lastFolderPath = filepath.Join(picsDir, folders[len(folders)-2].Name())
	} else {
		lastFolderPath = filepath.Join(picsDir, folders[len(folders)-1].Name())
	}
	pics, err := ioutil.ReadDir(lastFolderPath)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while reading pictures directory %s", err)
		return
	}
	for _, fileInfo := range pics {
		picPath := filepath.Join(lastFolderPath, fileInfo.Name())
		resp.Files = append(resp.Files, picPath)
	}
	resp.Hash, err = makeHash(pics, lastFolderPath)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SetAdv - получает gzip json, в котором указаны имена и тела файлов, закодированных в base64
// проводит раскодирование и записывает полученные файлы в static/images, удаляя старое содержимое
func SetAdv(w http.ResponseWriter, r *http.Request) {
	var entries []FileEntry

	re, err := gzip.NewReader(r.Body)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Can't create gzip reader %s", err)
		return
	}
	defer re.Close()
	decoder := json.NewDecoder(re)
	if err = decoder.Decode(&entries); err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Can't decode entries json %s", err)
		return
	}

	tempDirPath := filepath.Join(Config.ImagesRoot, "/temp/")
	if err = writeToTemp(&entries, tempDirPath); err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while handling directories %s", err)
		return
	}

	if err = handleTemp(&entries); err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("can't decode json file in SetAdv %s", err)
		return
	}
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
