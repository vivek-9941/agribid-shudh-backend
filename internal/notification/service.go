package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"github.com/agribid/agribid-shudh-backend/internal/integration"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Logger interface {
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
}

type Service struct {
	repo      Repository
	emailProv integration.EmailProvider
	smsProv   integration.SMSProvider
	logger    Logger
}

func NewService(repo Repository, emailProv integration.EmailProvider, smsProv integration.SMSProvider, logger Logger) *Service {
	return &Service{
		repo:      repo,
		emailProv: emailProv,
		smsProv:   smsProv,
		logger:    logger,
	}
}

func (s *Service) Send(ctx context.Context, eventType EventType, recipientID uuid.UUID, recipientEmail, recipientPhone string, data map[string]interface{}) {
	// Execute in background
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		pref, err := s.repo.GetPreferences(bgCtx, recipientID)
		if err != nil {
			s.logger.Error("failed to get preferences", zap.Error(err), zap.String("recipient", recipientID.String()))
			return
		}

		for _, muted := range pref.MutedEvents {
			if muted == string(eventType) {
				return // muted
			}
		}

		// Hardcoded templates for stub
		titleTmplStr := "Notification: {{.EventType}}"
		bodyTmplStr := "Event {{.EventType}} occurred."
		if eventType == EventOrderPlaced {
			titleTmplStr = "Order Placed"
			bodyTmplStr = "Order {{.OrderNumber}} has been placed successfully."
		}

		tmplData := data
		tmplData["EventType"] = string(eventType)

		titleTmpl, _ := template.New("title").Parse(titleTmplStr)
		bodyTmpl, _ := template.New("body").Parse(bodyTmplStr)

		var titleBuf, bodyBuf bytes.Buffer
		_ = titleTmpl.Execute(&titleBuf, tmplData)
		_ = bodyTmpl.Execute(&bodyBuf, tmplData)

		title := titleBuf.String()
		body := bodyBuf.String()

		if pref.InAppEnabled {
			dataBytes, _ := json.Marshal(data)
			notif := &Notification{
				RecipientID: recipientID,
				Type:        string(eventType),
				Title:       title,
				Body:        body,
				DataPayload: dataBytes,
			}
			_ = s.repo.Create(bgCtx, notif)
		}

		if pref.EmailEnabled && recipientEmail != "" && s.emailProv != nil {
			s.sendWithRetry(bgCtx, func() error {
				return s.emailProv.SendEmail(bgCtx, recipientEmail, title, body)
			}, "email")
		}

		if pref.SMSEnabled && recipientPhone != "" && s.smsProv != nil {
			s.sendWithRetry(bgCtx, func() error {
				return s.smsProv.SendSMS(bgCtx, recipientPhone, body)
			}, "sms")
		}
	}()
}

func (s *Service) sendWithRetry(ctx context.Context, fn func() error, channel string) {
	maxRetries := 3
	backoff := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		err := fn()
		if err == nil {
			return // Success
		}

		if i == maxRetries-1 {
			s.logger.Error(fmt.Sprintf("%s delivery failed after retries", channel), zap.Error(err))
			return
		}
		
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
			backoff *= 2
		}
	}
}

func (s *Service) ListForUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Notification, int64, error) {
	return s.repo.ListForUser(ctx, userID, page, pageSize)
}

func (s *Service) MarkRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.MarkRead(ctx, id, userID)
}

func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreference, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *Service) UpsertPreferences(ctx context.Context, pref *NotificationPreference) error {
	return s.repo.UpsertPreferences(ctx, pref)
}
