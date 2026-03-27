package components

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"strings"
	"sync"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	xdraw "golang.org/x/image/draw"
)

var albumArtCache sync.Map

// RenderAlbumArt converts embedded cover art into ANSI truecolor half-blocks.
// width is measured in terminal cells and rows is measured in terminal rows.
func RenderAlbumArt(data []byte, width, rows int) string {
	if width < 4 || rows < 2 {
		return ""
	}

	if len(data) == 0 {
		return ""
	}

	format := artFormat()
	cacheKey := fmt.Sprintf("%x:%dx%d:%s", md5.Sum(data), width, rows, format)
	if cached, ok := albumArtCache.Load(cacheKey); ok {
		return cached.(string)
	}

	if rendered, err := renderWithChafa(data, width, rows, format); err == nil && rendered != "" {
		albumArtCache.Store(cacheKey, rendered)
		return rendered
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return ""
	}

	target := image.NewRGBA(image.Rect(0, 0, width, rows*2))
	xdraw.CatmullRom.Scale(target, target.Bounds(), img, img.Bounds(), xdraw.Over, nil)

	var sb strings.Builder
	for y := 0; y < target.Bounds().Dy(); y += 2 {
		for x := 0; x < target.Bounds().Dx(); x++ {
			top := target.At(x, y)
			bottom := top
			if y+1 < target.Bounds().Dy() {
				bottom = target.At(x, y+1)
			}
			sb.WriteString(ansiHalfBlock(top, bottom))
		}
		sb.WriteString("\x1b[0m")
		if y+2 < target.Bounds().Dy() {
			sb.WriteByte('\n')
		}
	}

	rendered := sb.String()
	albumArtCache.Store(cacheKey, rendered)
	return rendered
}

func ansiHalfBlock(top, bottom color.Color) string {
	tr, tg, tb, _ := top.RGBA()
	br, bg, bb, _ := bottom.RGBA()
	return fmt.Sprintf(
		"\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀",
		uint8(tr>>8), uint8(tg>>8), uint8(tb>>8),
		uint8(br>>8), uint8(bg>>8), uint8(bb>>8),
	)
}

func renderWithChafa(data []byte, width, rows int, format string) (string, error) {
	if os.Getenv("MUSIC_PLAYER_DISABLE_CHAFA") == "1" {
		return "", fmt.Errorf("chafa disabled")
	}

	chafaPath, err := exec.LookPath("chafa")
	if err != nil {
		return "", err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	tempFile, err := os.CreateTemp("", "gtmpc-cover-*.png")
	if err != nil {
		return "", err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)
	defer tempFile.Close()

	if err := png.Encode(tempFile, img); err != nil {
		return "", err
	}

	args := []string{
		"--format=" + format,
		"--size", fmt.Sprintf("%dx%d", width, rows),
		"--animate=off",
		"--optimize=5",
		"--relative=off",
		"--colors=full",
		tempPath,
	}
	if format == "symbols" {
		args = append(args[:len(args)-1], "--symbols=block+border+half+space+wide", tempPath)
	}

	cmd := exec.Command(chafaPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(output), "\n"), nil
}

func artFormat() string {
	if override := os.Getenv("MUSIC_PLAYER_ART_FORMAT"); override != "" {
		return override
	}
	return "symbols"
}
