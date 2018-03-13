package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"spMDOImages/server/conf"
	gzip "spUtils/gzippool"
)

type JResp struct {
	Hash  string   `json:"hash"`
	Files []string `json:"file"`
}

type FileEntry struct {
	Filename  string `json:"filename"`
	Content64 string `json:"file"`
}

func RegisterRoutes(mux *http.ServeMux) *http.ServeMux {
	GetHandler := http.HandlerFunc(GetAdv)
	SetHandler := http.HandlerFunc(SetAdv)
	mux.Handle("/api/get-adv", GetRequestMiddle(AuthMiddle(GetHandler)))
	mux.Handle("/api/set-adv", PostRequestMiddle(AuthMiddle(SetHandler)))
	return mux
}

// GetAdv отправляет клиенту json, в котором указан список файлов в папке static/images и их хеш
func GetAdv(w http.ResponseWriter, r *http.Request) {
	var resp JResp

	picturePaths, err := getFilePaths()
	if err != nil {
		http.Error(w, "Internal service error", http.StatusInternalServerError)
		log.Printf("Couldn't read picture %s", err)
		return
	}
	for _, p := range picturePaths {
		url := conf.ServerURL + p
		fmt.Println("PATH: ", url)
		resp.Files = append(resp.Files, url)
	}
	fmt.Printf("FILE PATHS: %+v", picturePaths)
	resp.Hash, err = makeHash(picturePaths)
	if err != nil {
		http.Error(w, "Internal service error", http.StatusInternalServerError)
		log.Printf("Couldn't calculate hash %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	respB, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Internal service error", http.StatusInternalServerError)
		log.Printf("Couldn't marshal response %s", err)
		return
	}
	_, err = w.Write(respB)
	if err != nil {
		http.Error(w, "Internal service error", http.StatusInternalServerError)
		log.Printf("Couldn't write a response %s", err)
		return
	}
}

// SetAdv - получает gzip json, в котором указаны имена и тела файлов, закодированных в base64
// проводит раскодирование и записывает полученные файлы в static/images, удаляя старое содержимое
func SetAdv(w http.ResponseWriter, r *http.Request) {
	var entries []FileEntry

	gReader, err := gzip.NewReader(r.Body)
	if err != nil {
		http.Error(w, "Internal service error", http.StatusInternalServerError)
		log.Printf("Couldn't read request body %s", err)
		return
	}

	if err = json.NewDecoder(gReader).Decode(&entries); err != nil {
		http.Error(w, "Internal service error", http.StatusInternalServerError)
		log.Printf("Can't decode entries json %s", err)
		return
	}

	if err = writeFiles(&entries); err != nil {
		http.Error(w, "Internal service error", http.StatusInternalServerError)
		log.Printf("Error while writing files -> %s", err)
		return
	}
}
