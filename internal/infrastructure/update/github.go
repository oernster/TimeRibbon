// Package update answers the application's ReleaseSource from GitHub's latest-release endpoint, the
// one network request TimeRibbon makes (NFR-S-1, FR-509). The endpoint answers only a published
// release that is neither a draft nor a prerelease, so a tag pushed during development is never
// seen: the guard is the endpoint's own contract, not a check here. The HTTP client is injected, so
// the tests never touch the network. Ported from PigeonPost.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/oernster/timeribbon/internal/application"
)

// LatestReleaseAPIURL is GitHub's latest-release endpoint for this repository.
const LatestReleaseAPIURL = "https://api.github.com/repos/oernster/TimeRibbon/releases/latest"

// acceptHeader asks the GitHub API for its JSON.
const acceptHeader = "application/vnd.github+json"

// requestTimeout bounds the one request, so a check never waits long.
const requestTimeout = 5 * time.Second

// maxBody caps what is read of the answer: a release's JSON is a few kilobytes, so anything near
// this is not one. A size the server states is never trusted.
const maxBody = 1 << 20

// Doer sends one HTTP request; *http.Client is one and the tests stand in another.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// releasePayload is the part of GitHub's latest-release JSON the check reads.
type releasePayload struct {
	TagName string         `json:"tag_name"`
	HTMLURL string         `json:"html_url"`
	Assets  []assetPayload `json:"assets"`
}

type assetPayload struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

// GitHub is an application.ReleaseSource over the GitHub API.
type GitHub struct {
	apiURL string
	client Doer
}

// New answers the production source, with a client that gives up after requestTimeout.
func New() *GitHub {
	return NewWith(LatestReleaseAPIURL, &http.Client{Timeout: requestTimeout})
}

// NewWith answers a source asking apiURL through client.
func NewWith(apiURL string, client Doer) *GitHub {
	return &GitHub{apiURL: apiURL, client: client}
}

// LatestRelease answers the latest published release; an error when it cannot be read, which the
// service treats as no answer.
func (g *GitHub) LatestRelease(ctx context.Context) (application.ReleaseInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.apiURL, nil)
	if err != nil {
		return application.ReleaseInfo{}, fmt.Errorf("building the release request: %w", err)
	}
	req.Header.Set("Accept", acceptHeader)
	resp, err := g.client.Do(req)
	if err != nil {
		return application.ReleaseInfo{}, fmt.Errorf("asking for the latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return application.ReleaseInfo{}, fmt.Errorf("asking for the latest release: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return application.ReleaseInfo{}, fmt.Errorf("reading the release: %w", err)
	}
	var payload releasePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return application.ReleaseInfo{}, fmt.Errorf("reading the release: %w", err)
	}
	if payload.TagName == "" || payload.HTMLURL == "" {
		return application.ReleaseInfo{}, fmt.Errorf("the release names no version or no page")
	}
	assets := make([]application.ReleaseAsset, 0, len(payload.Assets))
	for _, asset := range payload.Assets {
		if asset.Name != "" && asset.DownloadURL != "" {
			assets = append(assets, application.ReleaseAsset{Name: asset.Name, DownloadURL: asset.DownloadURL})
		}
	}
	return application.ReleaseInfo{Version: payload.TagName, PageURL: payload.HTMLURL, Assets: assets}, nil
}
