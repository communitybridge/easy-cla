// Copyright The Linux Foundation and each contributor to CommunityBridge.
// SPDX-License-Identifier: MIT

package docraptor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/communitybridge/easycla/cla-backend-go/logging"
)

var (
	errInvalidKey = errors.New("invalid key")
)

const (
	docraptorURL = "https://%s@docraptor.com/docs"
)

// Client structure model
type Client struct {
	apiKey   string
	url      string
	testMode bool
}

// NewDocraptorClient creates a new docraptor client instance
func NewDocraptorClient(key string, testMode bool) (Client, error) {
	if key == "" {
		return Client{}, errInvalidKey
	}

	url := fmt.Sprintf(docraptorURL, key)

	return Client{
		apiKey:   key,
		url:      url,
		testMode: testMode,
	}, nil
}

// CreatePDF accepts an HTML document and returns a PDF
func (dc Client) CreatePDF(html string) (io.ReadCloser, error) {
	document := map[string]interface{}{
		"document_type":    "pdf",
		"document_content": html,
		"name":             "docraptor-go.pdf",
		"test":             dc.testMode,
	}

	documentBytes, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}

	log.Debug("Generating PDF using docraptor...")
	resp, err := http.Post(dc.url, "application/json", bytes.NewBuffer(documentBytes))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
