package ocr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/log"
)

// PaddleBillOCRRawItem represents one raw recognized transaction item returned by PaddleOCR HTTP service.
// It matches the JSON structure provided by the external Paddle service.
type PaddleBillOCRRawItem struct {
	Amount   string `json:"amount"`
	Classify string `json:"classify"`
	Account  string `json:"account"`
	Date     string `json:"date"`
	Project  string `json:"project"`
	Label    string `json:"label"`
	Text     string `json:"text"`
}

// PaddleBillOCRResponse represents the response body of the PaddleOCR HTTP service.
// Expected JSON structure:
//
//	{
//	  "success": true,
//	  "raw": [
//	    {
//	      "amount": "-123.45",
//	      "classify": "食品饮料-食品",
//	      "account": "微信",
//	      "date": "2026-02-11 22:31:24",
//	      "project": "项目1",
//	      "label": "标签1",
//	      "text": "2月7日 21:49 京东超市 -100.00"
//	    }
//	  ]
//	}
type PaddleBillOCRResponse struct {
	Success bool                 `json:"success"`
	Raw     []PaddleBillOCRRawItem `json:"raw"`
	Error   string               `json:"error,omitempty"`
}

// RunPaddleBillOCR sends the image data to an external PaddleOCR HTTP service and returns the raw recognized items.
// The external service is responsible for doing OCR and returning a JSON body like the structure documented above.
func RunPaddleBillOCR(imageData []byte, endpoint string) ([]PaddleBillOCRRawItem, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("image data is empty")
	}
	if endpoint == "" {
		return nil, fmt.Errorf("paddle ocr endpoint is not configured")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("image", "bill.jpg")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err := part.Write(imageData); err != nil {
		return nil, fmt.Errorf("write image data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, &body)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Warnf(nil, "[ocr.RunPaddleBillOCR] request paddle ocr endpoint failed: %v", err)
		return nil, fmt.Errorf("request paddle ocr endpoint failed: %w", err)
	}
	defer resp.Body.Close()

	var ocrResp PaddleBillOCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrResp); err != nil {
		log.Warnf(nil, "[ocr.RunPaddleBillOCR] decode paddle ocr response failed: %v", err)
		return nil, fmt.Errorf("decode paddle ocr response failed: %w", err)
	}

	if !ocrResp.Success {
		if ocrResp.Error != "" {
			return nil, fmt.Errorf("paddle ocr failed: %s", ocrResp.Error)
		}
		return nil, fmt.Errorf("paddle ocr failed without error message")
	}

	if len(ocrResp.Raw) == 0 {
		return nil, fmt.Errorf("paddle ocr returned no raw items")
	}

	return ocrResp.Raw, nil
}
