package ocr

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/log"
)

const (
	// TesseractLangChineseSimplified and English for bill/screenshot text
	TesseractLangBill = "chi_sim+eng"
)

// RunTesseract runs tesseract OCR on image data and returns recognized text.
// Requires tesseract to be installed (e.g. apt install tesseract-ocr tesseract-ocr-chi-sim).
// If tesseract is not available, returns an error that can be checked with errs.IsTesseractNotAvailable.
func RunTesseract(imageData []byte, lang string) (string, error) {
	if len(imageData) == 0 {
		return "", fmt.Errorf("image data is empty")
	}
	if lang == "" {
		lang = TesseractLangBill
	}

	// Determine image extension from content or default to png
	ext := getImageExtension(imageData)
	tmpDir := os.TempDir()
	inputPath := filepath.Join(tmpDir, "ezbk_ocr_in"+ext)
	outputBase := filepath.Join(tmpDir, "ezbk_ocr_out")

	if err := os.WriteFile(inputPath, imageData, 0600); err != nil {
		return "", fmt.Errorf("write temp image: %w", err)
	}
	defer os.Remove(inputPath)
	defer os.Remove(outputBase + ".txt")

	cmd := exec.Command("tesseract", inputPath, outputBase, "-l", lang)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Warnf(nil, "[ocr.RunTesseract] tesseract failed: %v, stderr: %s", err, stderr.String())
		return "", fmt.Errorf("tesseract: %w (is tesseract installed? e.g. tesseract-ocr, tesseract-ocr-chi-sim)", err)
	}

	out, err := os.ReadFile(outputBase + ".txt")
	if err != nil {
		return "", fmt.Errorf("read tesseract output: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func getImageExtension(data []byte) string {
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 {
		return ".jpg"
	}
	if len(data) >= 8 && bytes.Equal(data[0:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return ".png"
	}
	return ".png"
}
