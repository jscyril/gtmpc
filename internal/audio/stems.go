package audio

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/jscyril/golang_music_player/api"
	"github.com/jscyril/golang_music_player/internal/logger"
)

// StemCache resolves alternate playback files for each track/mode combination.
// It prefers Demucs source separation when available and falls back to ffmpeg
// channel filtering otherwise.
type StemCache struct {
	root         string
	mu           sync.Mutex
	demucsCmd    []string
	demucsModel  string
	demucsShifts string
}

const stemAlgorithmVersion = "v4"

func NewStemCache(root string) *StemCache {
	return &StemCache{
		root:         root,
		demucsCmd:    detectDemucsCommand(),
		demucsModel:  envOrDefault("MUSIC_PLAYER_DEMUCS_MODEL", "htdemucs_ft"),
		demucsShifts: envOrDefault("MUSIC_PLAYER_DEMUCS_SHIFTS", "4"),
	}
}

func (c *StemCache) Resolve(track *api.Track, mode api.AudioMode) (string, error) {
	if track == nil {
		return "", fmt.Errorf("nil track")
	}
	if mode == api.ModeNormal {
		return track.FilePath, nil
	}

	if err := os.MkdirAll(c.root, 0o755); err != nil {
		return "", fmt.Errorf("create stem cache: %w", err)
	}

	output := filepath.Join(c.root, c.fileName(track, mode))
	if _, err := os.Stat(output); err == nil {
		return output, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat stem file: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := os.Stat(output); err == nil {
		return output, nil
	}

	if err := c.generate(track, output, mode); err != nil {
		return "", err
	}
	return output, nil
}

func (c *StemCache) fileName(track *api.Track, mode api.AudioMode) string {
	sum := md5.Sum([]byte(track.FilePath))
	return fmt.Sprintf("%x-%s-%s.wav", sum[:8], mode.String(), stemAlgorithmVersion)
}

func (c *StemCache) generate(track *api.Track, outputPath string, mode api.AudioMode) error {
	if err := c.generateWithDemucs(track); err == nil {
		if _, err := os.Stat(outputPath); err == nil {
			return nil
		}
		return fmt.Errorf("demucs did not produce %s output", mode.String())
	} else if err != errDemucsUnavailable {
		return err
	}

	return c.generateWithFFmpeg(track.FilePath, outputPath, mode)
}

func (c *StemCache) generateWithFFmpeg(inputPath, outputPath string, mode api.AudioMode) error {
	filter, err := ffmpegFilter(mode)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", inputPath,
		"-vn",
		"-ac", "2",
		"-af", filter,
		"-c:a", "pcm_s16le",
		outputPath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("generate %s stem: %w (%s)", mode.String(), err, string(output))
	}

	return nil
}

var errDemucsUnavailable = fmt.Errorf("demucs backend unavailable")

func (c *StemCache) generateWithDemucs(track *api.Track) error {
	if len(c.demucsCmd) == 0 {
		return errDemucsUnavailable
	}

	vocalsPath := filepath.Join(c.root, c.fileName(track, api.ModeVocals))
	karaokePath := filepath.Join(c.root, c.fileName(track, api.ModeKaraoke))
	if fileExists(vocalsPath) && fileExists(karaokePath) {
		return nil
	}

	sum := md5.Sum([]byte(track.FilePath))
	workspace := filepath.Join(c.root, ".demucs-work", fmt.Sprintf("%x", sum[:8]))
	outputDir := filepath.Join(workspace, "out")
	cacheDir := filepath.Join(c.root, ".demucs-cache")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create demucs output dir: %w", err)
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return fmt.Errorf("create demucs cache dir: %w", err)
	}
	defer os.RemoveAll(workspace)

	trackName := trimExtension(filepath.Base(track.FilePath))
	args := append([]string{}, c.demucsCmd[1:]...)
	args = append(args,
		"-n", c.demucsModel,
		"--two-stems=vocals",
		"--shifts", c.demucsShifts,
		"-d", "cpu",
		"-o", outputDir,
		"--filename", "{track}-{stem}.{ext}",
		track.FilePath,
	)

	cmd := exec.Command(c.demucsCmd[0], args...)
	cmd.Env = append(os.Environ(),
		"XDG_CACHE_HOME="+cacheDir,
		"TORCH_HOME="+filepath.Join(cacheDir, "torch"),
	)

	logger.Info("Running Demucs separation for %q", track.Title)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("run demucs: %w (%s)", err, string(output))
	}

	modelDir := filepath.Join(outputDir, c.demucsModel)
	vocalsSrc := filepath.Join(modelDir, fmt.Sprintf("%s-vocals.wav", trackName))
	karaokeSrc := filepath.Join(modelDir, fmt.Sprintf("%s-no_vocals.wav", trackName))
	if !fileExists(vocalsSrc) || !fileExists(karaokeSrc) {
		return fmt.Errorf("demucs output missing expected stems")
	}

	if err := moveOrCopyFile(vocalsSrc, vocalsPath); err != nil {
		return fmt.Errorf("store vocals stem: %w", err)
	}
	if err := moveOrCopyFile(karaokeSrc, karaokePath); err != nil {
		return fmt.Errorf("store karaoke stem: %w", err)
	}

	return nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func detectDemucsCommand() []string {
	if python := os.Getenv("MUSIC_PLAYER_DEMUCS_PYTHON"); python != "" && fileExists(python) {
		return []string{python, "-m", "demucs"}
	}

	localPython := filepath.Join(".tools", "demucs-venv", "bin", "python")
	if fileExists(localPython) {
		return []string{localPython, "-m", "demucs"}
	}

	if demucsPath, err := exec.LookPath("demucs"); err == nil {
		return []string{demucsPath}
	}

	return nil
}

func moveOrCopyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func trimExtension(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name
	}
	return name[:len(name)-len(ext)]
}

func ffmpegFilter(mode api.AudioMode) (string, error) {
	switch mode {
	case api.ModeKaraoke:
		// Build a stronger karaoke estimate by deriving an enhanced center
		// channel and subtracting it back out of the stereo image.
		return "aformat=channel_layouts=stereo,dialoguenhance=original=1:enhance=2.2:voice=10,pan=stereo|c0=0.9*FL-1.1*FC|c1=0.9*FR-1.1*FC,stereotools=mode=lr>lr:mlev=0.03125:slev=1.08:base=0,highpass=f=120,lowpass=f=14000,dynaudnorm=f=250:g=15", nil
	case api.ModeVocals:
		// Extract the enhanced center channel and shape it toward vocal range.
		return "aformat=channel_layouts=stereo,dialoguenhance=original=0:enhance=2.6:voice=12,pan=stereo|c0=FC|c1=FC,highpass=f=120,lowpass=f=8000,dynaudnorm=f=200:g=9", nil
	default:
		return "", fmt.Errorf("unsupported audio mode: %s", mode.String())
	}
}
