package components

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"sync"
	"testing"
)

func TestRenderAlbumArtNoData(t *testing.T) {
	t.Setenv("MUSIC_PLAYER_DISABLE_CHAFA", "1")
	rendered := RenderAlbumArt(nil, 8, 4)
	if rendered != "" {
		t.Fatal("expected no art output when no embedded thumbnail is available")
	}
}

func TestRenderAlbumArtImage(t *testing.T) {
	t.Setenv("MUSIC_PLAYER_DISABLE_CHAFA", "1")
	albumArtCache = sync.Map{}
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})
	img.Set(0, 1, color.RGBA{0, 0, 255, 255})
	img.Set(1, 1, color.RGBA{255, 255, 0, 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	rendered := RenderAlbumArt(buf.Bytes(), 6, 3)
	if rendered == "" {
		t.Fatal("expected image art to render")
	}
	if !strings.Contains(rendered, "▀") {
		t.Fatal("expected image art to use half-block rendering")
	}
}
