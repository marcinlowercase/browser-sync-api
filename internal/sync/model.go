package sync

type SyncPayload struct {
	Timestamp int64            `json:"timestamp"`
	Profiles  []ProfileSyncDTO `json:"profiles"`
}

type ProfileSyncDTO struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Settings    ProfileSettingsSyncDTO `json:"settings"`
	PinnedApps  []AppSyncDTO           `json:"pinnedApps"`
	VisitedURLs []VisitedUrlSyncDTO    `json:"visitedUrls"`
}

type ProfileSettingsSyncDTO struct {
	DefaultURL            string  `json:"defaultUrl"`
	AnimationSpeed        float32 `json:"animationSpeed"`
	IsSharpMode           bool    `json:"isSharpMode"`
	CursorContainerSize   float32 `json:"cursorContainerSize"`
	CursorPointerSize     float32 `json:"cursorPointerSize"`
	CursorTrackingSpeed   float32 `json:"cursorTrackingSpeed"`
	ShowSuggestions       bool    `json:"showSuggestions"`
	ClosedTabHistorySize  float32 `json:"closedTabHistorySize"`
	BackSquareOffsetX     float32 `json:"backSquareOffsetX"`
	BackSquareOffsetY     float32 `json:"backSquareOffsetY"`
	BackSquareIdleOpacity float32 `json:"backSquareIdleOpacity"`
	SearchEngine          int     `json:"searchEngine"`
	IsFullscreenMode      bool    `json:"isFullscreenMode"`
	HighlightColor        int     `json:"highlightColor"`
	IsAdBlockEnabled      bool    `json:"isAdBlockEnabled"`
	IsGuideModeEnabled    bool    `json:"isGuideModeEnabled"`
	IsDesktopMode         bool    `json:"isDesktopMode"`
	IsEnabledMediaControl bool    `json:"isEnabledMediaControl"`
	IsEnabledOutSync      bool    `json:"isEnabledOutSync"`
	OptionsOrder          string  `json:"optionsOrder"`
	SettingsOrder         string  `json:"settingsOrder"`
	HiddenOptions         string  `json:"hiddenOptions"`
}

type AppSyncDTO struct {
	ID      int64  `json:"id"`
	Label   string `json:"label"`
	URL     string `json:"url"`
	IconURL string `json:"iconUrl"`
}

type VisitedUrlSyncDTO struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}
