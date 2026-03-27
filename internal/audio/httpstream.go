// Package audio — httpstream.go provides HTTP-based audio streaming.
// NewHTTPStreamer fetches audio from an HTTP endpoint using an Authorization header
// and returns a beep.StreamSeekCloser + format for integration with the audio engine.
//
// This is used by the client mode (cmd/client) to stream audio from the server.
// It does NOT touch the local file-based AudioEngine; PlayFromURL is added separately.
package audio

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/wav"
)

// httpReadSeekCloser wraps an HTTP response body to provide io.ReadSeekCloser.
// Seeking is implemented by making new HTTP Range requests.
type httpReadSeekCloser struct {
	url      string
	token    string
	client   *http.Client
	resp     *http.Response
	position int64
	size     int64
}

// newHTTPReadSeekCloser opens an HTTP connection to url and returns the body.
func newHTTPReadSeekCloser(url, token string) (*httpReadSeekCloser, error) {
	c := &http.Client{} // no timeout for streaming
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return nil, fmt.Errorf("server responded %d for stream URL", resp.StatusCode)
	}

	return &httpReadSeekCloser{
		url:      url,
		token:    token,
		client:   c,
		resp:     resp,
		position: 0,
		size:     resp.ContentLength,
	}, nil
}

func (h *httpReadSeekCloser) Read(p []byte) (int, error) {
	n, err := h.resp.Body.Read(p)
	h.position += int64(n)
	return n, err
}

// Seek re-opens the HTTP connection with a Range header for the new position.
func (h *httpReadSeekCloser) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = h.position + offset
	case io.SeekEnd:
		if h.size < 0 {
			return 0, fmt.Errorf("cannot seek from end: unknown content length")
		}
		abs = h.size + offset
	default:
		return 0, fmt.Errorf("invalid seek whence %d", whence)
	}
	if abs < 0 {
		abs = 0
	}

	h.resp.Body.Close()

	req, err := http.NewRequest("GET", h.url, nil)
	if err != nil {
		return 0, fmt.Errorf("range request build: %w", err)
	}
	if h.token != "" {
		req.Header.Set("Authorization", "Bearer "+h.token)
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-", abs))

	resp, err := h.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("range request: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return 0, fmt.Errorf("server returned %d on range request", resp.StatusCode)
	}

	h.resp = resp
	h.position = abs
	return abs, nil
}

func (h *httpReadSeekCloser) Close() error {
	return h.resp.Body.Close()
}

// contentType returns the cleaned MIME type from a Content-Type header value.
func contentType(header string) string {
	parts := strings.SplitN(header, ";", 2)
	return strings.ToLower(strings.TrimSpace(parts[0]))
}

// NewHTTPStreamer opens an HTTP audio stream and returns a beep.StreamSeekCloser.
//
// Format detection is done via the Content-Type response header:
//   - audio/mpeg, audio/mp3  -> MP3
//   - audio/wav, audio/x-wav -> WAV
//   - audio/flac, audio/x-flac -> FLAC
//   - audio/ogg -> unsupported (graceful error)
//   - unknown -> MP3 fallback
func NewHTTPStreamer(url string, token string) (beep.StreamSeekCloser, beep.Format, error) {
	body, err := newHTTPReadSeekCloser(url, token)
	if err != nil {
		return nil, beep.Format{}, fmt.Errorf("open http stream: %w", err)
	}

	ct := contentType(body.resp.Header.Get("Content-Type"))

	switch ct {
	case "audio/mpeg", "audio/mp3":
		s, f, err := mp3.Decode(body)
		if err != nil {
			body.Close()
			return nil, beep.Format{}, fmt.Errorf("mp3 decode: %w", err)
		}
		return s, f, nil

	case "audio/wav", "audio/x-wav", "audio/wave":
		s, f, err := wav.Decode(body)
		if err != nil {
			body.Close()
			return nil, beep.Format{}, fmt.Errorf("wav decode: %w", err)
		}
		return s, f, nil

	case "audio/flac", "audio/x-flac":
		s, f, err := flac.Decode(body)
		if err != nil {
			body.Close()
			return nil, beep.Format{}, fmt.Errorf("flac decode: %w", err)
		}
		return s, f, nil

	case "audio/ogg", "application/ogg":
		body.Close()
		return nil, beep.Format{}, fmt.Errorf("OGG/Vorbis is not supported by this client; the server should provide MP3, WAV, or FLAC")

	default:
		// Unknown content type — attempt MP3 as a safe fallback
		s, f, err := mp3.Decode(body)
		if err != nil {
			body.Close()
			return nil, beep.Format{}, fmt.Errorf("unknown content-type %q, mp3 fallback failed: %w", ct, err)
		}
		return s, f, nil
	}
}
