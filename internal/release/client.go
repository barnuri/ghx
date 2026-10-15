package release

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	apiBase        = "https://api.github.com/repos/"
	requestTimeout = 2 * time.Minute
	maxAssetBytes  = 200 << 20
)

type Client struct {
	http *http.Client
}

func NewClient() Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	return Client{http: &http.Client{Timeout: requestTimeout, Transport: transport}}
}

func (client Client) Latest(repo string) (Release, error) {
	body, err := client.get(apiBase + repo + "/releases/latest")
	if err != nil {
		return Release{}, err
	}
	var latest Release
	if err := json.Unmarshal(body, &latest); err != nil {
		return Release{}, fmt.Errorf("decode latest release of %s: %w", repo, err)
	}
	return latest, nil
}

// DownloadVerified downloads asset and checks it against the sha256 line in checksums.
func (client Client) DownloadVerified(asset Asset, checksums Asset) ([]byte, error) {
	sums, err := client.get(checksums.DownloadURL)
	if err != nil {
		return nil, err
	}
	expected, err := ChecksumFor(string(sums), asset.Name)
	if err != nil {
		return nil, err
	}
	content, err := client.get(asset.DownloadURL)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(content)
	if actual := hex.EncodeToString(digest[:]); actual != expected {
		return nil, fmt.Errorf("checksum mismatch for %s: expected %s, got %s", asset.Name, expected, actual)
	}
	return content, nil
}

func ChecksumFor(checksums string, fileName string) (string, error) {
	for _, line := range strings.Split(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == fileName {
			return strings.ToLower(fields[0]), nil
		}
	}
	return "", fmt.Errorf("no checksum listed for %s", fileName)
}

func (client Client) get(url string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "ghx")
	response, err := client.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, response.StatusCode)
	}
	return io.ReadAll(io.LimitReader(response.Body, maxAssetBytes))
}
