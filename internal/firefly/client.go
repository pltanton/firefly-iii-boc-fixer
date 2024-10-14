package firefly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pltanton/firefly-iii-boc-fixer/internal/config"
)

type FireflyClient struct {
	client *http.Client
	c      config.Config
}

func NewClient(c config.Config) *FireflyClient {

	return &FireflyClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		c: c,
	}
}

func (fc *FireflyClient) UpdateTransaction(id int64, req TransactionUpdateRequest) error {
	url, err := url.JoinPath(fc.c.FireflyURL, "/api/v1/transactions", strconv.FormatInt(id, 10))
	if err != nil {
		return fmt.Errorf("failed to build url: %w", url)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}
	fc.enrichDefault(httpReq)

	resp, err := fc.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("Failed to read body", "err", err)
		}

		slog.Error("Non success response received", "status", resp.StatusCode, "body", string(respBody), "url", url)
		return fmt.Errorf("non OK status code returned")
	}

	return nil
}

func (fc *FireflyClient) enrichDefault(req *http.Request) {
	bearer := "Bearer " + fc.c.FireflyToken

	req.Header.Add("Authorization", bearer)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/vnd.api+json")
}

type TransactionUpdateRequest struct {
	ApplyRules   bool                     `json:"apply_rules"`
	FireWebhooks bool                     `json:"fire_webhooks"`
	GroupTitle   *string                  `json:"group_title,omitempty"`
	Transactions []TransactionSplitUpdate `json:"transactions"`
}

type TransactionSplitUpdate struct {
	TransactionJournalId int64       `json:"transaction_journal_id"`
	Description          string      `json:"description,omitempty"`
	Date                 **time.Time `json:"date,omitempty"`
	ProcessDate          **time.Time `json:"process_date,omitempty"`
	PaymentDate          **time.Time `json:"payment_date,omitempty"`
	Tags                 []string    `json:"tags,omitempty"`
}

func TimeAtMidday(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, time.UTC)
}

func Time(t time.Time) **time.Time {
	tPtr := &t
	return &tPtr
}
