package main

import (
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

type JResp struct {
	Hash  string   `json:"hash"`
	Files []string `json:"file"`
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
		fmt.Println(err)
	}
	req.Header.Set("Token", token)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return false, err
	}
	if resp.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}

func GetAdv(w http.ResponseWriter, r *http.Request) {
	authorized, err := isAuthorized(r)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if authorized == false {
		http.Error(w, "Authorization required", http.StatusUnauthorized)
		log.Println(err)
		return
	}

	workDir, err := filepath.Abs(".")
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	var resp JResp
	picsDir := filepath.Join(workDir, "./static/images")
	pics, err := ioutil.ReadDir(picsDir)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	for _, fileInfo := range pics {
		picPath := filepath.Join(picsDir, fileInfo.Name())
		resp.Files = append(resp.Files, picPath)
	}
	resp.Hash = "YET TO IMPLEMENT"
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func SetAdv(w http.ResponseWriter, r *http.Request) {
	authorized, err := isAuthorized(r)
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if authorized == false {
		http.Error(w, "Authorization required", http.StatusUnauthorized)
		log.Println(err)
		return
	}

	var entries []FileEntry
	picsDir, err := filepath.Abs("./static/images")
	if err != nil {
		http.Error(w, "Internal service problems", http.StatusInternalServerError)
		log.Println(err)
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
		log.Println("Decode error", err)
		return
	}

	fmt.Printf("%+v", entries)
	for _, entry := range entries {
		path := filepath.Join(picsDir, entry.Filename)
		content, err := base64.StdEncoding.DecodeString(entry.Content64)
		if err != nil {
			http.Error(w, "Internal service problems", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		ioutil.WriteFile(path, content, 0744)
	}
}
