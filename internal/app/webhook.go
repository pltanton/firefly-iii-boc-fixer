package app

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"

	"github.com/pltanton/firefly-iii-boc-fixer/internal/config"
	"github.com/pltanton/firefly-iii-boc-fixer/internal/firefly"
)

func handleWebhook(
	c config.Config,
	fireflyClient *firefly.FireflyClient,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request, err := validateAndParseWHRequest(r, c.Secret)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			slog.Warn("Received message with wrong or un-parsable signature!", "err", err)
			return
		}

		slog.Debug("Request received", "request", request)

		if err = fixBOC(request, fireflyClient); err != nil {
			slog.Error("Failed to fix BOC message", "id", request.Content.ID, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

// Validate signature and parse request body to struct
func validateAndParseWHRequest(r *http.Request, secret []byte) (WebhookRequest, error) {
	rawRequest, err := io.ReadAll(r.Body)
	if err != nil {
		return WebhookRequest{}, fmt.Errorf("failed to read body: %w", err)
	}

	providedSign, err := parseSignature(r.Header.Get("Signature"))
	if err != nil {
		return WebhookRequest{}, fmt.Errorf("failed to parse signature: %w", err)
	}

	expectedSing := buildSignature(providedSign.Timestamp, rawRequest, secret)
	if !hmac.Equal(providedSign.Sign, expectedSing) {
		return WebhookRequest{}, fmt.Errorf("signature doesn't match")
	}

	var parsed WebhookRequest
	if err = json.Unmarshal(rawRequest, &parsed); err != nil {
		return WebhookRequest{}, fmt.Errorf("failed to parse request json: %w", err)
	}

	return parsed, nil
}

// Build signature from timestamp and raw message
func buildSignature(ts int64, rawRequest, secret []byte) []byte {
	mac := hmac.New(sha3.New256, secret)
	mac.Write([]byte(strconv.FormatInt(ts, 10)))
	mac.Write([]byte{'.'})
	mac.Write(rawRequest)

	return mac.Sum(nil)
}

// Parses signature
func parseSignature(raw string) (signature, error) {
	var result signature
	for _, kv := range strings.Split(raw, ",") {
		split := strings.Split(kv, "=")
		if len(split) != 2 {
			return signature{}, fmt.Errorf("unexpected kv part of signature: %s", kv)
		}
		switch split[0] {
		case "t":
			if result.Timestamp != 0 {
				return signature{}, fmt.Errorf("wrong signature structure, multiple t")
			}
			t, err := strconv.ParseInt(split[1], 10, 64)
			if err != nil {
				return signature{}, fmt.Errorf("failed to parse timestamp '%s': %w", split[1], err)
			}
			result.Timestamp = t
		case "v1":
			if result.Sign != nil {
				return signature{}, fmt.Errorf("wrong signature structure, multiple v1")
			}
			var err error
			result.Sign, err = hex.DecodeString(split[1])
			if err != nil {
				return signature{}, fmt.Errorf("failed to decode signature value '%s' as hex string: %w", split[1], err)
			}
		default:
			slog.Warn("unknown part of signature", "kv", kv)
		}
	}

	if result.Timestamp == 0 || result.Sign == nil {
		return signature{}, fmt.Errorf("signature '%s' incomplete", raw)
	}

	return result, nil
}

type signature struct {
	Timestamp int64
	Sign      []byte
}
