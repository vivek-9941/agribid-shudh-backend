package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

type SMSProvider interface {
	SendSMS(ctx context.Context, to string, message string) error
}

type EmailProvider interface {
	SendEmail(ctx context.Context, to string, subject string, body string) error
}

type StorageProvider interface {
	Upload(ctx context.Context, filename string, reader io.Reader) (string, error)
	GetURL(filename string) (string, error)
}

type Logger interface {
	Info(msg string, fields ...zap.Field)
}

type MockSMSProvider struct {
	logger Logger
}

func NewMockSMSProvider(logger Logger) SMSProvider {
	return &MockSMSProvider{logger: logger}
}

func (m *MockSMSProvider) SendSMS(ctx context.Context, to string, message string) error {
	m.logger.Info("Mock SMS Sent", zap.String("to", to), zap.String("message", message))
	return nil
}

type MockEmailProvider struct {
	logger Logger
}

func NewMockEmailProvider(logger Logger) EmailProvider {
	return &MockEmailProvider{logger: logger}
}

func (m *MockEmailProvider) SendEmail(ctx context.Context, to string, subject string, body string) error {
	m.logger.Info("Mock Email Sent", zap.String("to", to), zap.String("subject", subject))
	return nil
}

type LocalStorageProvider struct {
	baseDir string
	baseURL string
}

func NewLocalStorageProvider(baseDir, baseURL string) (StorageProvider, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &LocalStorageProvider{
		baseDir: baseDir,
		baseURL: baseURL,
	}, nil
}

func (l *LocalStorageProvider) Upload(ctx context.Context, filename string, reader io.Reader) (string, error) {
	// Add timestamp to ensure uniqueness
	timestamp := time.Now().Format("20060102150405")
	uniqueFilename := fmt.Sprintf("%s_%s", timestamp, filename)
	
	path := filepath.Join(l.baseDir, uniqueFilename)
	
	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	if _, err := io.Copy(file, reader); err != nil {
		return "", err
	}
	
	return l.GetURL(uniqueFilename)
}

func (l *LocalStorageProvider) GetURL(filename string) (string, error) {
	return fmt.Sprintf("%s/uploads/%s", l.baseURL, filename), nil
}

type MockPDFGenerator struct{}

func NewMockPDFGenerator() *MockPDFGenerator {
	return &MockPDFGenerator{}
}

func (m *MockPDFGenerator) GenerateInvoicePDF(inv interface{}) (string, error) {
	return "http://localhost:8080/uploads/dummy_invoice.pdf", nil
}
