package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

type Handler struct {
	DB *sql.DB
}

func generate9DigitCode() (string, error) {
	max := big.NewInt(1000000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%09d", n.Int64()), nil
}

// RequestCode handles POST /api/v1/auth/request-code
func (h *Handler) RequestCode(w http.ResponseWriter, r *http.Request) {
	var payload RequestCodePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("❌ [AUTH/REQUEST] Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("📩 [AUTH/REQUEST] Code requested for email: %s", payload.Email)

	code, err := generate9DigitCode()
	if err != nil {
		log.Printf("❌ [AUTH/REQUEST] Failed to generate code for %s | Error: %v", payload.Email, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	expiration := time.Now().Add(10 * time.Minute)
	query := `
		INSERT INTO otp_codes (email, code, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET code = EXCLUDED.code, expires_at = EXCLUDED.expires_at;`

	_, err = h.DB.Exec(query, payload.Email, code, expiration)
	if err != nil {
		log.Printf("❌ [AUTH/REQUEST] DB Error for %s | Error: %v", payload.Email, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	go sendAuthEmail(payload.Email, code)

	log.Printf("✅[AUTH/REQUEST] Code generation successful for: %s", payload.Email)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Message: "Code sent successfully"})
}

// VerifyCode handles POST /api/v1/auth/verify-code
func (h *Handler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	var payload VerifyCodePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("❌ [AUTH/VERIFY] Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("🔐 [AUTH/VERIFY] Verification attempt for email: %s", payload.Email)

	var dbCode string
	var expiresAt time.Time
	err := h.DB.QueryRow(`SELECT code, expires_at FROM otp_codes WHERE email = $1`, payload.Email).Scan(&dbCode, &expiresAt)

	if err == sql.ErrNoRows || dbCode != payload.Code || time.Now().After(expiresAt) {
		log.Printf("❌[AUTH/VERIFY] Invalid or expired code for: %s", payload.Email)
		http.Error(w, "Invalid or expired code", http.StatusUnauthorized)
		return
	}

	h.DB.Exec(`DELETE FROM otp_codes WHERE email = $1`, payload.Email)

	var userID string
	err = h.DB.QueryRow(`
		INSERT INTO users (email) VALUES ($1)
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		RETURNING id;`, payload.Email).Scan(&userID)

	if err != nil {
		log.Printf("❌ [AUTH/VERIFY] DB Error creating user %s | Error: %v", payload.Email, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	token, err := GenerateJWT(userID, payload.Email)
	if err != nil {
		log.Printf("❌ [AUTH/VERIFY] Failed to generate token for %s | Error: %v", payload.Email, err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [AUTH/VERIFY] Login successful for email: %s (UserID: %s)", payload.Email, userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, Message: "Login successful"})
}

func sendAuthEmail(toEmail string, code string) {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	subject := "Subject: the browser of oo1 studio login code\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body style="text-align: center;">
			<div style="font-family: monospace; font-size: 16px">
				%s
			</div>
		</body>
		</html>`, code)
	message := []byte(subject + mime + htmlBody)

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)

	if err != nil {
		log.Printf("❌ [EMAIL] Failed to send email to %s: %v", toEmail, err)
		return
	}

	log.Printf("📧 [EMAIL] Real email successfully sent to %s", toEmail)
}
