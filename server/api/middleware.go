package api

import (
	"fmt"
	"log"
	"net/http"
	"spMDOImages/server/conf"
	"time"
)

func isAuthorized(r *http.Request) (auth bool, err error) {
	token := r.Header.Get("Token")
	client := &http.Client{Timeout: 10 * time.Second}
	url := conf.VerifyTokenURL
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("Error while creating request for auth %s", err)
	}
	req.Header.Set("Token", token)
	resp, err := client.Do(req)
	fmt.Println(token)
	if err != nil {
		return false, fmt.Errorf("Error while making request for auth %s", err)
	}
	if resp.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}

func AuthMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorized, err := isAuthorized(r)
		if err != nil {
			http.Error(w, "Internal service problems", http.StatusInternalServerError)
			log.Printf("can't authorize %s", err)
			return
		}
		if authorized == false {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			log.Printf("user wasn't authorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetRequestMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Use GET method", http.StatusMethodNotAllowed)
			log.Printf("Bad request method on %s", r.RequestURI)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func PostRequestMiddle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Use POST method", http.StatusMethodNotAllowed)
			log.Printf("Bad request method on %s", r.RequestURI)
			return
		}
		next.ServeHTTP(w, r)
	})
}
