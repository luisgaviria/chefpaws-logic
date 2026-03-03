package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/luisgaviria/chefpaws-logic/internal/email"
)

// ContactRequest matches the JSON payload sent by the ContactForm component.
type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Breed   string `json:"breed"`
	Message string `json:"message"`
}

// NewContactHandler returns an http.HandlerFunc that emails the team on contact form submission.
// Passing an empty resendKey skips email gracefully (useful in local dev without credentials).
func NewContactHandler(resendKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ContactRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Server-side validation (mirrors client-side Zod schema)
		req.Name = strings.TrimSpace(req.Name)
		req.Email = strings.TrimSpace(req.Email)
		req.Message = strings.TrimSpace(req.Message)

		if req.Name == "" || req.Email == "" || req.Message == "" {
			jsonError(w, "Name, email, and message are required", http.StatusBadRequest)
			return
		}
		if !strings.Contains(req.Email, "@") {
			jsonError(w, "Invalid email address", http.StatusBadRequest)
			return
		}

		fmt.Printf("📩 CONTACT: %s <%s> breed=%q\n", req.Name, req.Email, req.Breed)

		if resendKey != "" {
			if err := email.SendContactEmail(resendKey, req.Name, req.Email, req.Breed, req.Message); err != nil {
				fmt.Printf("⚠️  Contact email failed: %v\n", err)
				jsonError(w, "Failed to send message — please try again", http.StatusInternalServerError)
				return
			}
			fmt.Printf("✅ Contact email sent for %s\n", req.Email)
		} else {
			fmt.Println("⚠️  RESEND_API_KEY not set — contact email skipped")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}

// jsonError writes a JSON {"error":"..."} response with the given status code.
func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
