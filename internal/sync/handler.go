package sync

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"browser-sync-api/internal/auth"
)

type Handler struct {
	DB *sql.DB
}

// POST /api/v1/sync/push
func (h *Handler) PushData(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	log.Printf("⬆️ [PUSH] Started for UserID: %s", userID)

	var payload SyncPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("❌ [PUSH] Invalid JSON payload from UserID: %s | Error: %v", userID, err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	var userExists bool
	err := h.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&userExists)
	if err != nil || !userExists {
		log.Printf("❌ [PUSH] User no longer exists for UserID: %s", userID)
		http.Error(w, "User account no longer exists. Please log out and log in again.", http.StatusUnauthorized)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		log.Printf("❌ [PUSH] Failed to begin transaction for UserID: %s | Error: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	log.Printf("🧹 [PUSH] Wiping old data for UserID: %s", userID)
	_, err = tx.Exec(`DELETE FROM sync_profiles WHERE user_id = $1`, userID)
	if err != nil {
		log.Printf("❌ [PUSH] Failed to clear old sync data for UserID: %s | Error: %v", userID, err)
		http.Error(w, "Failed to clear old sync data", http.StatusInternalServerError)
		return
	}

	seenProfiles := make(map[string]bool)

	for _, profile := range payload.Profiles {
		if seenProfiles[profile.ID] {
			log.Printf("⚠️ [PUSH] Skipping duplicate profile ID: %s", profile.ID)
			continue
		}
		seenProfiles[profile.ID] = true

		var dbProfileID string
		err = tx.QueryRow(`
			INSERT INTO sync_profiles (user_id, client_profile_id, name)
			VALUES ($1, $2, $3) RETURNING id`,
			userID, profile.ID, profile.Name).Scan(&dbProfileID)

		if err != nil {
			log.Printf("❌ [PUSH] Error inserting profile %s: %v", profile.ID, err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(`
					INSERT INTO sync_profile_settings (profile_id, settings_json)
					VALUES ($1, $2)`,
			dbProfileID, profile.Settings)

		if err != nil {
			log.Printf("❌ [PUSH] Error inserting settings for profile %s: %v", profile.ID, err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		for _, app := range profile.PinnedApps {
			_, err = tx.Exec(`INSERT INTO sync_pinned_apps (profile_id, client_app_id, label, url, icon_url) VALUES ($1, $2, $3, $4, $5)`,
				dbProfileID, app.ID, app.Label, app.URL, app.IconURL)
			if err != nil {
				log.Printf("⚠️ [PUSH] Error inserting app %s: %v", app.Label, err)
			}
		}

		for _, url := range profile.VisitedURLs {
			_, err = tx.Exec(`INSERT INTO sync_visited_urls (profile_id, url, title) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
				dbProfileID, url.URL, url.Title)
			if err != nil {
				log.Printf("⚠️ [PUSH] Error inserting url %s: %v", url.URL, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("❌ [PUSH] Failed to commit transaction for UserID: %s | Error: %v", userID, err)
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [PUSH] Successful for UserID: %s (Saved %d profiles)", userID, len(payload.Profiles))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Sync successful"}`))
}

// GET /api/v1/sync/pull
func (h *Handler) PullData(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	log.Printf("⬇️ [PULL] Started for UserID: %s", userID)

	payload := SyncPayload{
		Timestamp: time.Now().UnixMilli(),
		Profiles:  []ProfileSyncDTO{},
	}

	profileRows, err := h.DB.Query(`SELECT id, client_profile_id, name FROM sync_profiles WHERE user_id = $1`, userID)
	if err != nil {
		log.Printf("❌ [PULL] Failed to query profiles for UserID: %s | Error: %v", userID, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer profileRows.Close()

	for profileRows.Next() {
		var dbID, clientID, name string
		profileRows.Scan(&dbID, &clientID, &name)

		profile := ProfileSyncDTO{
			ID:          clientID,
			Name:        name,
			PinnedApps:  []AppSyncDTO{},
			VisitedURLs: []VisitedUrlSyncDTO{},
		}

		var settingsJson string
		err = h.DB.QueryRow(`SELECT settings_json FROM sync_profile_settings WHERE profile_id = $1`, dbID).Scan(&settingsJson)

		if err == nil {
			profile.Settings = settingsJson
		}

		appRows, _ := h.DB.Query(`SELECT client_app_id, label, url, icon_url FROM sync_pinned_apps WHERE profile_id = $1`, dbID)
		for appRows.Next() {
			var a AppSyncDTO
			appRows.Scan(&a.ID, &a.Label, &a.URL, &a.IconURL)
			profile.PinnedApps = append(profile.PinnedApps, a)
		}
		appRows.Close()

		urlRows, _ := h.DB.Query(`SELECT url, title FROM sync_visited_urls WHERE profile_id = $1`, dbID)
		for urlRows.Next() {
			var u VisitedUrlSyncDTO
			urlRows.Scan(&u.URL, &u.Title)
			profile.VisitedURLs = append(profile.VisitedURLs, u)
		}
		urlRows.Close()

		payload.Profiles = append(payload.Profiles, profile)
	}

	log.Printf("✅ [PULL] Successful for UserID: %s (Returned %d profiles)", userID, len(payload.Profiles))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

// DELETE /api/v1/sync/account
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	log.Printf("🗑️ [DELETE] Started for UserID: %s", userID)

	_, err := h.DB.Exec(`DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		log.Printf("❌ [DELETE] Failed to delete user account %s | Error: %v", userID, err)
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [DELETE] Successful for UserID: %s", userID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Account deleted successfully"}`))
}
