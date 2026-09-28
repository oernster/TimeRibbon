package update

import (
	"context"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/timeribbon/internal/application"
)

// fakeDoer answers every request with status and body (with err instead when err is set); it keeps
// the request it saw.
type fakeDoer struct {
	status int
	body   io.Reader
	err    error
	seen   *http.Request
}

func (f *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	f.seen = req
	if f.err != nil {
		return nil, f.err
	}
	return &http.Response{StatusCode: f.status, Body: io.NopCloser(f.body)}, nil
}

func answering(status int, body string) *fakeDoer {
	return &fakeDoer{status: status, body: strings.NewReader(body)}
}

// failingReader fails every read, as a connection dropped mid-answer does.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("connection dropped") }

const release = `{"tag_name":"v2.1.0","html_url":"https://example.test/r","assets":[
	{"name":"TimeRibbonSetup.exe","browser_download_url":"https://example.test/a.exe"},
	{"name":"","browser_download_url":"https://example.test/nameless"},
	{"name":"TimeRibbon.dmg","browser_download_url":""},
	{"name":"timeribbon.flatpak","browser_download_url":"https://example.test/a.flatpak"}]}`

func TestTheLatestReleaseIsReadWithOnlyWholeAssets(t *testing.T) {
	t.Parallel()
	doer := answering(http.StatusOK, release)
	got, err := NewWith(LatestReleaseAPIURL, doer).LatestRelease(context.Background())
	want := []application.ReleaseAsset{
		{Name: "TimeRibbonSetup.exe", DownloadURL: "https://example.test/a.exe"},
		{Name: "timeribbon.flatpak", DownloadURL: "https://example.test/a.flatpak"},
	}
	if err != nil || got.Version != "v2.1.0" || got.PageURL != "https://example.test/r" || !slices.Equal(got.Assets, want) {
		t.Fatalf("answered %+v, %v", got, err)
	}
	if doer.seen.URL.String() != LatestReleaseAPIURL || doer.seen.Header.Get("Accept") != acceptHeader || doer.seen.Method != http.MethodGet {
		t.Errorf("asked %s %s with Accept %q", doer.seen.Method, doer.seen.URL, doer.seen.Header.Get("Accept"))
	}
}

func TestTheProductionSourceAsksThisRepositoryAndGivesUp(t *testing.T) {
	t.Parallel()
	source := New()
	client, ok := source.client.(*http.Client)
	if source.apiURL != LatestReleaseAPIURL || !ok || client.Timeout != requestTimeout {
		t.Errorf("asks %s through %#v", source.apiURL, source.client)
	}
	if !strings.Contains(LatestReleaseAPIURL, "/oernster/TimeRibbon/releases/latest") {
		t.Errorf("%s is not this repository's latest release", LatestReleaseAPIURL)
	}
}

func TestEveryUnusableAnswerIsAnError(t *testing.T) {
	t.Parallel()
	cases := map[string]*fakeDoer{
		"unreachable":    {err: errors.New("no route")},
		"not found":      answering(http.StatusNotFound, release),
		"dropped":        {status: http.StatusOK, body: failingReader{}},
		"not JSON":       answering(http.StatusOK, "<html>"),
		"no tag":         answering(http.StatusOK, `{"html_url":"https://example.test/r"}`),
		"no page":        answering(http.StatusOK, `{"tag_name":"v2.1.0"}`),
		"wrong types":    answering(http.StatusOK, `{"tag_name":7,"html_url":"https://example.test/r"}`),
		"oversized":      answering(http.StatusOK, `{"tag_name":"v2.1.0","html_url":"`+strings.Repeat("x", maxBody)+`"}`),
		"not an object":  answering(http.StatusOK, `[]`),
		"assets unusual": answering(http.StatusOK, `{"tag_name":"v2.1.0","html_url":"x","assets":{}}`),
	}
	for name, doer := range cases {
		if got, err := NewWith(LatestReleaseAPIURL, doer).LatestRelease(context.Background()); err == nil {
			t.Errorf("%s answered %+v with no error", name, got)
		}
	}
}

func TestAnAddressThatCannotBeAskedIsAnError(t *testing.T) {
	t.Parallel()
	if _, err := NewWith("://", answering(http.StatusOK, release)).LatestRelease(context.Background()); err == nil {
		t.Error("a malformed address was asked")
	}
}
