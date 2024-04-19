package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.ServeFile(w, r, "index.html")
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.ServeFile(w, r, "index.html")
		case http.MethodPost:
			handlePostRequest(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var requestBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Get the user's request
	userRequest := requestBody["request"].(string)

	// Call the OpenAI API
	apiKey := ""
	apiEndpoint := "https://api.openai.com/v1/completions"
	client := resty.New()
	requestBody["messages"] = []interface{}{map[string]interface{}{"role": "system", "content": userRequest}}
	requestBody["max_tokens"] = 512

	response, err := client.R().
		SetAuthToken(apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(requestBody).
		Post(apiEndpoint)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Fatalf("Error while sending the request: %v", err)
		return
	}

	// Parse the response
	var data map[string]interface{}
	if err := json.Unmarshal(response.Body(), &data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Error while decoding JSON response:", err)
		return
	}

	// Extract the content from the JSON response
	content := data["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

	// Write the response to the client
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, content)
}
