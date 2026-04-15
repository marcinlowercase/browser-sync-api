package sync

type SyncPayload struct {
	Timestamp int64            `json:"timestamp"`
	Profiles  []ProfileSyncDTO `json:"profiles"`
}

type ProfileSyncDTO struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Settings    string              `json:"settings"`
	PinnedApps  []AppSyncDTO        `json:"pinnedApps"`
	VisitedURLs []VisitedUrlSyncDTO `json:"visitedUrls"`
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
