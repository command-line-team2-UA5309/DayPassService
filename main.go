package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"time"
)

type dayPass struct {
	DayCode string `json:"daycode"`
}

func codeGen() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		log.Println("daily code error:", err)
	}
	return hex.EncodeToString(b)
}

func codeRefresh() {
	for {
		newCode := codeGen()

		f, err := os.Create("./passcode.txt")
		if err != nil {
			log.Println("file can't be created:", err)
		}

		_, err = f.Write([]byte(newCode))
		if err != nil {
			log.Println("file can't be updated", err)
		}

		f.Close()

		current := time.Now()
		nextChange := time.Date(current.Year(), current.Month(), current.Day()+1, 0, 0, 0, 0, current.Location())
		time.Sleep(nextChange.Sub(current))
	}
}

func checkCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req dayPass
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	f, err := os.Open("./passcode.txt")
	if err != nil {
		log.Println("file opening error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	bs, err := io.ReadAll(f)
	if err != nil {
		log.Println("file reading error:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	currentCode := string(bs)

	w.Header().Set("Content-Type", "application/json")
	if req.DayCode == currentCode {
		err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		if err != nil {
			log.Println("JSON Encoding error:", err)
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		err := json.NewEncoder(w).Encode(map[string]string{"status": "error"})
		if err != nil {
			log.Println("JSON Encoding error:", err)
		}
	}
}

func main() {
	go codeRefresh()

	http.HandleFunc("/check-code", checkCodeHandler)

	log.Fatal(http.ListenAndServe(":9090", nil))
}
