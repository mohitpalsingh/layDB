package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RequestPayLoad struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ResponseJson struct {
	Status  string `json:"key"`
	Message string `json:"value"`
}

func handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		var rp RequestPayLoad

		err = json.Unmarshal(body, &rp)
		if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
		}

		err = db.Set(rp.Key, rp.Value)
		if err != nil {
			responseJSON(w, ResponseJson{
				Status:  "success",
				Message: "Key Value pair saved successfully.",
			}, http.StatusOK)
		}
	} else {
		fmt.Println("Invalid request method.")
		fmt.Fprintf(w, "Invalid request method.")
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		key := r.URL.Query().Get("key")
		value, err := db.Get(key)
		if err != nil {
			responseJSON(w, ResponseJson{
				Status:  "error",
				Message: err.Error(),
			}, http.StatusNotFound)
			return
		}

		responseJSON(w, RequestPayLoad{
			Key:   key,
			Value: value,
		}, http.StatusOK)
	} else {
		fmt.Println("Invalid request method.")
		fmt.Fprintf(w, "Invalid request method.")
	}
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		key := r.URL.Query().Get("key")
		err := db.Delete(key)
		if err != nil {
			responseJSON(w, ResponseJson{
				Status:  "error",
				Message: err.Error(),
			}, http.StatusInternalServerError)
			return
		}

		responseJSON(w, ResponseJson{
			Status:  "success",
			Message: "Key deleted successfully.",
		}, http.StatusOK)
	} else {
		fmt.Println("Invalid request method.")
		fmt.Fprintf(w, "Invalid request method.")
	}
}

func responseJSON(w http.ResponseWriter, data interface{}, status int) {
	d, err := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	w.WriteHeader(status)
	w.Write(d)
}

var db *LayDB

func main() {
	db, err := NewDb(&Config{
		FilePath:   "/tmp/layDB",
		FileData:   "/database.txt",
		DeleteData: "/database_delete.txt",
	})
	if err != nil {
		fmt.Println("Error opening db: ", err)
		return
	}
	defer db.Close()

	go db.CompactFile()
	go db.DeleteFromFile()

	http.HandleFunc("/set", handleSet)
	http.HandleFunc("/get", handleGet)
	http.HandleFunc("/delete", handleDelete)

	address := ":8000"

	fmt.Printf("Server is running on localhost:%s", address)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		fmt.Println("Error starting the server: ", err)
		return
	}

}
