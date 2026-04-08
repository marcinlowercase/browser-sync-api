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
	// Extract the user ID that the Auth Middleware put into the context
	userID := r.Context().Value(auth.UserIDKey).(string)

	var payload SyncPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Begin Database Transaction
	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Safely rollback if we don't call tx.Commit()

	// 1. Wipe old data: Because of ON DELETE CASCADE, deleting profiles deletes apps, history, & settings!
	_, err = tx.Exec(`DELETE FROM sync_profiles WHERE user_id = $1`, userID)
	if err != nil {
		http.Error(w, "Failed to clear old sync data", http.StatusInternalServerError)
		return
	}

	// 2. Insert new data from the payload
	for _, profile := range payload.Profiles {
		var dbProfileID string
		err = tx.QueryRow(`
			INSERT INTO sync_profiles (user_id, client_profile_id, name)
			VALUES ($1, $2, $3) RETURNING id`,
			userID, profile.ID, profile.Name).Scan(&dbProfileID)

		if err != nil {
			log.Printf("Error inserting profile: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Insert Settings
		s := profile.Settings
		_, err = tx.Exec(`
			INSERT INTO sync_profile_settings (
				profile_id, default_url, animation_speed, is_sharp_mode, cursor_container_size,
				cursor_pointer_size, cursor_tracking_speed, show_suggestions, closed_tab_history_size,
				back_square_offset_x, back_square_offset_y, back_square_idle_opacity, search_engine,
				is_fullscreen_mode, highlight_color, is_ad_block_enabled, is_guide_mode_enabled,
				is_desktop_mode, is_enabled_media_control, is_enabled_out_sync, options_order,
				settings_order, hidden_options
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
			)`,
			dbProfileID, s.DefaultURL, s.AnimationSpeed, s.IsSharpMode, s.CursorContainerSize,
			s.CursorPointerSize, s.CursorTrackingSpeed, s.ShowSuggestions, s.ClosedTabHistorySize,
			s.BackSquareOffsetX, s.BackSquareOffsetY, s.BackSquareIdleOpacity, s.SearchEngine,
			s.IsFullscreenMode, s.HighlightColor, s.IsAdBlockEnabled, s.IsGuideModeEnabled,
			s.IsDesktopMode, s.IsEnabledMediaControl, s.IsEnabledOutSync, s.OptionsOrder,
			s.SettingsOrder, s.HiddenOptions)

		if err != nil {
			log.Printf("Error inserting settings: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Insert Pinned Apps
		for _, app := range profile.PinnedApps {
			_, err = tx.Exec(`INSERT INTO sync_pinned_apps (profile_id, client_app_id, label, url, icon_url) VALUES ($1, $2, $3, $4, $5)`,
				dbProfileID, app.ID, app.Label, app.URL, app.IconURL)
			if err != nil {
				log.Printf("Error inserting app: %v", err)
			}
		}

		// Insert History
		for _, url := range profile.VisitedURLs {
			_, err = tx.Exec(`INSERT INTO sync_visited_urls (profile_id, url, title) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
				dbProfileID, url.URL, url.Title)
			if err != nil {
				log.Printf("Error inserting url: %v", err)
			}
		}
	}

	// Commit Transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Sync successful"}`))
}

// GET /api/v1/sync/pull
func (h *Handler) PullData(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)

	payload := SyncPayload{
		Timestamp: time.Now().UnixMilli(),
		Profiles:  []ProfileSyncDTO{},
	}

	// Get Profiles
	profileRows, err := h.DB.Query(`SELECT id, client_profile_id, name FROM sync_profiles WHERE user_id = $1`, userID)
	if err != nil {
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

		// Get Settings
		var s ProfileSettingsSyncDTO
		err = h.DB.QueryRow(`
			SELECT default_url, animation_speed, is_sharp_mode, cursor_container_size,
			cursor_pointer_size, cursor_tracking_speed, show_suggestions, closed_tab_history_size,
			back_square_offset_x, back_square_offset_y, back_square_idle_opacity, search_engine,
			is_fullscreen_mode, highlight_color, is_ad_block_enabled, is_guide_mode_enabled,
			is_desktop_mode, is_enabled_media_control, is_enabled_out_sync, options_order,
			settings_order, hidden_options
			FROM sync_profile_settings WHERE profile_id = $1`, dbID).Scan(
			&s.DefaultURL, &s.AnimationSpeed, &s.IsSharpMode, &s.CursorContainerSize,
			&s.CursorPointerSize, &s.CursorTrackingSpeed, &s.ShowSuggestions, &s.ClosedTabHistorySize,
			&s.BackSquareOffsetX, &s.BackSquareOffsetY, &s.BackSquareIdleOpacity, &s.SearchEngine,
			&s.IsFullscreenMode, &s.HighlightColor, &s.IsAdBlockEnabled, &s.IsGuideModeEnabled,
			&s.IsDesktopMode, &s.IsEnabledMediaControl, &s.IsEnabledOutSync, &s.OptionsOrder,
			&s.SettingsOrder, &s.HiddenOptions)

		if err == nil {
			profile.Settings = s
		}

		// Get Apps
		appRows, _ := h.DB.Query(`SELECT client_app_id, label, url, icon_url FROM sync_pinned_apps WHERE profile_id = $1`, dbID)
		for appRows.Next() {
			var a AppSyncDTO
			appRows.Scan(&a.ID, &a.Label, &a.URL, &a.IconURL)
			profile.PinnedApps = append(profile.PinnedApps, a)
		}
		appRows.Close()

		// Get History
		urlRows, _ := h.DB.Query(`SELECT url, title FROM sync_visited_urls WHERE profile_id = $1`, dbID)
		for urlRows.Next() {
			var u VisitedUrlSyncDTO
			urlRows.Scan(&u.URL, &u.Title)
			profile.VisitedURLs = append(profile.VisitedURLs, u)
		}
		urlRows.Close()

		payload.Profiles = append(payload.Profiles, profile)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

// DELETE /api/v1/sync/account
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)

	// Deleting the user automatically cascades and deletes all profiles, apps, history, and settings!
	_, err := h.DB.Exec(`DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		log.Printf("Failed to delete user account: %v", err)
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Account deleted successfully"}`))
}
