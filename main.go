package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
)

var baseURL = "https://xbl.io/api/v2/"

func main() {

	http.HandleFunc("/fetch", fetchHandler)
	http.HandleFunc("/friends.json", friendsJSONHandler)
	http.HandleFunc("/", chartHandler)

	log.Fatal(http.ListenAndServe(":8000", nil))
}

func fetchHandler(w http.ResponseWriter, r *http.Request) {
	publicKey := os.Getenv("PUBLIC_KEY")
	if publicKey == "" {
		publicKey = "default_public_key"
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+"friends", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Add("X-Authorization", publicKey)

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	people, ok := response["people"].([]interface{})
	if !ok {
		http.Error(w, "Invalid response structure", http.StatusInternalServerError)
		return
	}

	var friends []map[string]string
	for _, item := range people {
		friend, ok := item.(map[string]interface{})
		if !ok {
			http.Error(w, "Invalid friend structure", http.StatusInternalServerError)
			return
		}

		name, ok := friend["displayName"].(string)
		if !ok {
			http.Error(w, "Invalid displayName", http.StatusInternalServerError)
			return
		}

		score, ok := friend["gamerScore"].(string)
		if !ok {
			http.Error(w, "Invalid gamerScore", http.StatusInternalServerError)
			return
		}

		friends = append(friends, map[string]string{
			"Name":  name,
			"Score": score,
		})
	}

	friendsBytes, err := json.MarshalIndent(friends, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = ioutil.WriteFile("static/friends.json", friendsBytes, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func friendsJSONHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/friends.json")
}

type Friend struct {
	Name  string `json:"Name"`
	Score string `json:"Score"`
}

func chartHandler(w http.ResponseWriter, r *http.Request) {
	// Load the data from friends.json
	var friends []Friend
	data, err := ioutil.ReadFile("static/friends.json")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(data, &friends)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse the template
	tmpl, err := template.ParseFiles("static/chart.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template into a buffer
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, friends)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the buffer to the response
	w.Write(buf.Bytes())
}