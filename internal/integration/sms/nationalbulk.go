// Package sms provides the NationalBulkSMS gateway integration.
package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config holds all credentials / settings for the National Bulk SMS gateway.
type Config struct {
	APIURL       string // e.g. https://sms.nationalbulksms.com/fe/api/v1/send
	Username     string
	Password     string
	SenderID     string // DLT approved sender-id, e.g. KHAGLI
	DLTContentID string // 19-digit DLT content template ID
	DLTPEiD      string // 19-digit DLT principal entity ID
}

// Provider implements the auth.SMSProvider / integration.SMSProvider interface
// and sends real transactional SMS via National Bulk SMS gateway.
type Provider struct {
	cfg    Config
	client *http.Client
}

// New creates a new National Bulk SMS provider.
func New(cfg Config) *Provider {
	return &Provider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// apiResponse is the expected JSON response envelope from the gateway.
type apiResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SendSMS sends a transactional SMS to a 10-digit Indian mobile number.
// The `to` parameter may be in E.164 format (+91XXXXXXXXXX) or plain 10 digits.
func (p *Provider) SendSMS(ctx context.Context, to, message string) error {
	mobile := sanitizeMobile(to)
	if len(mobile) != 10 {
		return fmt.Errorf("sms: invalid mobile number %q – expected 10 digits after stripping country code", to)
	}

	params := url.Values{}
	params.Set("username", p.cfg.Username)
	params.Set("password", p.cfg.Password)
	params.Set("mobile", mobile)
	params.Set("message", message)
	params.Set("senderid", p.cfg.SenderID)
	params.Set("mtype", "N")      // Normal / Transactional
	params.Set("msgtype", "TEXT") // Plain text
	if p.cfg.DLTContentID != "" {
		params.Set("DLT_TE_ID", p.cfg.DLTContentID)
	}
	if p.cfg.DLTPEiD != "" {
		params.Set("PEID", p.cfg.DLTPEiD)
	}

	reqURL := fmt.Sprintf("%s?%s", p.cfg.APIURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("sms: build request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("sms: send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sms: gateway returned HTTP %d: %s", resp.StatusCode, body)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// Some gateways return plain text; treat 200 HTTP as success.
		return nil
	}

	// Code 200 / status "success" means delivered to gateway.
	statusLower := strings.ToLower(apiResp.Status)
	if statusLower != "success" && statusLower != "ok" && apiResp.Code != 200 {
		return fmt.Errorf("sms: gateway error – code=%d status=%s message=%s", apiResp.Code, apiResp.Status, apiResp.Message)
	}

	return nil
}

// sanitizeMobile strips the +91 or 91 country prefix and returns 10 raw digits.
func sanitizeMobile(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.TrimPrefix(phone, "+91")
	phone = strings.TrimPrefix(phone, "91")
	return phone
}
