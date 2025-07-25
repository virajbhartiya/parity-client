package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
	"github.com/theblitlabs/gologger"
	"github.com/theblitlabs/parity-client/internal/types"
)

// proxyHandler handles HTTP request proxying
type proxyHandler struct {
	serverURL   string
	deviceID    string
	creatorAddr string
	client      *http.Client
	logger      zerolog.Logger
}

// newProxyHandler creates a new proxy handler
func newProxyHandler(serverURL, deviceID, creatorAddr string) *proxyHandler {
	return &proxyHandler{
		serverURL:   strings.TrimSuffix(serverURL, "/"),
		deviceID:    deviceID,
		creatorAddr: creatorAddr,
		client:      &http.Client{},
		logger:      gologger.Get().With().Str("component", "proxy").Logger(),
	}
}

// forwardRequest forwards an HTTP request to the target server
func (p *proxyHandler) forwardRequest(w http.ResponseWriter, req *http.Request, path string) error {
	targetURL := fmt.Sprintf("%s/api/%s", p.serverURL, path)

	proxyReq, err := http.NewRequest(req.Method, targetURL, req.Body)
	if err != nil {
		return fmt.Errorf("error creating proxy request: %v", err)
	}

	types.CopyHeaders(proxyReq.Header, req.Header)

	proxyReq.Header.Set("X-Device-ID", p.deviceID)
	proxyReq.Header.Set("X-Creator-Address", p.creatorAddr)

	resp, err := p.client.Do(proxyReq)
	if err != nil {
		return fmt.Errorf("error forwarding request: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	types.CopyHeaders(w.Header(), resp.Header)

	w.WriteHeader(resp.StatusCode)

	if _, err := types.CopyBody(w, resp.Body); err != nil {
		p.logger.Error().Err(err).Msg("Failed to copy response body")
		return err
	}

	return nil
}
