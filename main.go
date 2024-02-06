package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"html/template"
	"net/http"
	"os"
	"strings"
	"strconv"
)


var tmpl = template.Must(template.ParseFiles("templates/index.tmpl"))
var limiter = rate.NewLimiter(1, 3)

type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func HomePageHandler(w http.ResponseWriter, r *http.Request, prefix string) {
	tmpl.Execute(w, map[string]interface{}{
		"doc_url": DocumentationURL,
		"prefix":  prefix,
	})
}

func getApiResponse(input string) ApiResponse {
	parts := strings.Split(input, ":")
	if len(parts) != 2 {
		return ApiResponse{Success: false, Message: "Invalid input! I'm expecting [host]:[port]"}
	}
	host, port := parts[0], parts[1]
	version, err := DetectMySQLServerVersion(host, port)
	if err != nil {
		switch e := err.(type) {
		case *ConnectionError:
			fmt.Fprintln(os.Stderr, "Failed to connect:", e.OriginalError)
			return ApiResponse{Success: false, Message: "I'm having trouble connecting to the server. Did you enter the correct host and port?"}
		case *ParseError:
			fmt.Fprintln(os.Stderr, "Failed to parse response:", e.OriginalError)
			return ApiResponse{Success: false, Message: "I didn't understand the server's response. Maybe there's something other than MySQL running on that port..."}
		default:
			fmt.Fprintln(os.Stderr, "An unexpected error occurred:", e)
			return ApiResponse{Success: false, Message: "Whoops, an unexpected error occurred! Please try again."}
		}
	}
	var successMessage = fmt.Sprintf("Nice! It looks like this server is running MySQL %s", version)
	return ApiResponse{Success: true, Message: successMessage}
}

func apiScanHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	if !limiter.Allow() {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return
	}
	clientIP := r.RemoteAddr
	query := r.URL.Query()
	input := query.Get("q")
	fmt.Println("New request from", clientIP)
	response := getApiResponse(input)
	// fmt.Println(response.Message)
	
	bytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshalling response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	responseStr := string(bytes)

	if err := LogRequestResponse(db, clientIP, input, responseStr); err != nil {
		fmt.Println("Error logging to database:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	
	if len(os.Args) == 4 && os.Args[1] == "cli" {
		host := os.Args[2]
		port := os.Args[3]
		version, err := DetectMySQLServerVersion(host, port)
		if err != nil {
			switch e := err.(type) {
			case *ConnectionError:
				fmt.Fprintln(os.Stderr, "Failed to connect:", e.OriginalError)
				fmt.Println("Failed to connect to the server.")
			case *ParseError:
				fmt.Fprintln(os.Stderr, "Failed to parse response:", e.OriginalError)
				fmt.Println("Failed to parse the server's response.")
			default:
				fmt.Fprintln(os.Stderr, "An unexpected error occurred:", e)
				fmt.Println("An unexpected error occurred.")
			}
			os.Exit(1)
		}
		fmt.Printf("MySQL server version: %s\n", version)
	} else if (len(os.Args) == 3 || len(os.Args) == 4) && os.Args[1] == "web" {
		port := os.Args[2]
		prefix := ""
		if len(os.Args) == 4 {
			prefix = os.Args[3]
		}
		_, err := strconv.Atoi(port)
		if err != nil {
			fmt.Println("Invalid port:", port)
			os.Exit(1)
		}
		db, err := InitDB()
		if err != nil {
			fmt.Println("Error initializing database:", err)
			os.Exit(1)
		}
		defer db.Close()
		http.HandleFunc("/"+prefix, func(w http.ResponseWriter, r *http.Request) {
			HomePageHandler(w, r, prefix)
		})
		apiPrefix := "/api/scan"
		if prefix != "" {
			apiPrefix = "/" + prefix + apiPrefix
		}
		http.HandleFunc(apiPrefix, func(w http.ResponseWriter, r *http.Request) {
			apiScanHandler(db, w, r)
		})
		fmt.Printf("Listening on port %s with prefix /%s ...\n", port, prefix)
		http.ListenAndServe(":"+port, nil)
	} else {
		fmt.Printf("Usage: %s cli <host> <port> | %s web <listen-port> [url-prefix]\n", os.Args[0], os.Args[0])
		os.Exit(1)
	}
}
