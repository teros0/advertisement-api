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

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/adv/get/", GetAdv).Methods("GET")
	router.HandleFunc("/api/adv/set/", SetAdv).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func isAuthorized(r *http.Request) (auth bool, err error) {
	token := r.Header.Get("Token")
	client := &http.Client{}
	url := "http://localhost:8000/api/auth/ver-token/"
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

	workDir, err := filepath.Abs(".")
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while getting absolute path %s", err)
		return
	}
	var resp JResp
	picsDir := filepath.Join(workDir, "./static/images")
	pics, err := ioutil.ReadDir(picsDir)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("Error while reading pictures directory %s", err)
		return
	}
	for _, fileInfo := range pics {
		picPath := filepath.Join(picsDir, fileInfo.Name())
		resp.Files = append(resp.Files, picPath)
	}
	resp.Hash, err = makeHash(pics, picsDir)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SetAdv - получает gzip json, в котором указаны имена и тела файлов, закодированных в base64
// проводит раскодирование и записывает полученные файлы в static/images, удаляя старое содержимое
func SetAdv(w http.ResponseWriter, r *http.Request) {
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

	var entries []FileEntry
	picsDir, err := filepath.Abs("./static/images")
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Printf("can't get absolute path in SetAdv %s", err)
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
		log.Printf("can't decode json file in SetAdv %s", err)
		return
	}

	if len(entries) == 0 {
		http.Error(w, "Specify at least one picture", http.StatusBadRequest)
		log.Printf("Agent tried to add 0 pictures")
		return
	}

	os.RemoveAll("./static/images")
	os.Mkdir("./static/images", 0744)

	fmt.Printf("Entries: %+v, Len of Entries %d", entries, len(entries))
	for _, entry := range entries {
		path := filepath.Join(picsDir, entry.Filename)
		content, err := base64.StdEncoding.DecodeString(entry.Content64)
		if err != nil {
			http.Error(w, "Internal service problems", http.StatusInternalServerError)
			log.Printf("error while encoding to base64 in SetAdv %s", err)
			return
		}
		ioutil.WriteFile(path, content, 0744)
	}
}

func makeHash(files []os.FileInfo, picsDir string) (string, error) {
	var allFiles []byte
	for _, file := range files {
		picPath := filepath.Join(picsDir, file.Name())
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
