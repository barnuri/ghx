package release

type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}
