package app

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/pltanton/firefly-iii-boc-fixer/internal/firefly"
)

const bocCountry = "CY"
const bocNumber = "0020"
const processedTag = "Processed by BoC fixer"

func isBOCIban(iban string) bool {
	return len(iban) > 8 && iban[0:2] == bocCountry && iban[4:8] == bocNumber
}

func fixBOC(
	request WebhookRequest,
	fireflyClient *firefly.FireflyClient,
) error {
	var logger = slog.Default().With("transaction-id", request.Content.ID)

	if len(request.Content.Transactions) != 1 {
		return nil
	}
	tx := request.Content.Transactions[0]

	// Skip not BoC transactions
	switch tx.Type {
	case "withdrawal":
		if !isBOCIban(tx.SourceIBAN) {
			logger.Debug("Skipping withdrawal by not BOC IBAN", "iban", tx.SourceIBAN)
			return nil
		}
	case "deposit":
		if !isBOCIban(tx.DestinationIBAN) {
			logger.Debug("Skipping deposit by not BOC IBAN", "iban", tx.DestinationIBAN)
			return nil
		}
	default:
		logger.Debug("Skipping transaction by type", "type", tx.Type)
		return nil
	}

	bocTx := parseBOCDescription(logger, tx.Description)
	newTags := []string{processedTag}
	if tx.Tags != nil && len(tx.Tags) > 0 {
		for _, tag := range tx.Tags {
			if tag == processedTag {
				logger.Debug("Skipping by presence of processing tag")
				return nil
			}
		}
		newTags = append(newTags, tx.Tags...)
	}

	if bocTx.Date.IsZero() {
		logger.Warn("Date not parsed from boc TX, skip", "id", request.Content.ID)
		return nil
	}

	logger.Info("BoC transaction received", "tx", bocTx)

	err := fireflyClient.UpdateTransaction(request.Content.ID, firefly.TransactionUpdateRequest{
		ApplyRules:   true,
		FireWebhooks: true,
		GroupTitle:   nil,
		Transactions: []firefly.TransactionSplitUpdate{
			{
				TransactionJournalId: request.Content.ID,
				Description:          bocTx.Description,
				Date:                 firefly.Time(firefly.TimeAtMidday(bocTx.Date)),
				PaymentDate:          firefly.Time(firefly.TimeAtMidday(bocTx.Date)),
				Tags:                 newTags,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	return nil
}

func parseBOCDescription(logger *slog.Logger, descriptionStr string) BOCTxByDescription {
	// Example: Card 5***3824 2024-10-10 71.41 EUR Auth 318622 Trace 357830 PURCHASE LU WWW.ALIEXPRESS.COM
	description := []byte(descriptionStr)

	var tx BOCTxByDescription

	// Parse amount with currency
	var amountRegex = regexp.MustCompile(`\d+\.+\d\d [A-Z]{3}`)
	if loc := amountRegex.FindIndex(description); loc != nil {
		tx.Amount = string(description[loc[0]:loc[1]])
		description = cutRespectSpace(description, loc[0], loc[1])
	} else {
		logger.Debug("Amount not found in BoC description", "description", descriptionStr)
	}

	// Parse Card
	var cardRegex = regexp.MustCompile(`Card ?\d\*{3}\d{4}`)
	if loc := cardRegex.FindIndex(description); loc != nil {
		tx.Card = strings.TrimSpace(string(description[loc[0]+4 : loc[1]]))
		description = cutRespectSpace(description, loc[0], loc[1])
	} else {
		logger.Debug("Card not found in BoC description", "description", descriptionStr)
	}

	// Parse Auth index
	var authRegex = regexp.MustCompile(`Auth ?\d+`)
	if loc := authRegex.FindIndex(description); loc != nil {
		tx.Auth = strings.TrimSpace(string(description[loc[0]+4 : loc[1]]))
		description = cutRespectSpace(description, loc[0], loc[1])
	} else {
		logger.Debug("Auth not found in BoC description", "description", descriptionStr)
	}

	// Parse Trace index
	var traceRegex = regexp.MustCompile(`Trace ?\d+`)
	if loc := traceRegex.FindIndex(description); loc != nil {
		tx.Trace = strings.TrimSpace(string(description[loc[0]+5 : loc[1]]))
		description = cutRespectSpace(description, loc[0], loc[1])
	} else {
		logger.Debug("Trace not found in BoC description", "description", descriptionStr)
	}

	// Parse transaction creation date
	var dateRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	if loc := dateRegex.FindIndex(description); loc != nil {
		dateStr := string(description[loc[0]:loc[1]])
		var err error
		tx.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			logger.Error("Failed to parse date from BoC description", "err", err, "description", descriptionStr)
		}
		description = cutRespectSpace(description, loc[0], loc[1])
	} else {
		logger.Debug("Date not found in BoC description", "description", descriptionStr)
	}

	var typeRegex = regexp.MustCompile(`PURCHASE|INWARD`)
	if loc := typeRegex.FindIndex(description); loc != nil {
		tx.Type = string(description[loc[0]:loc[1]])
		description = cutRespectSpace(description, loc[0], loc[1])
	} else {
		logger.Debug("Type not found in BoC description", "description", descriptionStr)
	}

	// Parse country
	var countryRegex = regexp.MustCompile(`^[A-Z]{2} | [A-Z]{2}$`)
	if loc := countryRegex.FindIndex(description); loc != nil {
		tx.Country = strings.TrimSpace(string(description[loc[0]:loc[1]]))
		description = cutRespectSpace(description, loc[0], loc[1])
	} else {
		logger.Debug("Country not found in BoC description", "description", descriptionStr)
	}

	tx.Description = string(description)

	return tx
}

func cutRespectSpace(slice []byte, l, r int) []byte {
	alreadyCut := false
	if l > 0 && slice[l-1] == ' ' {
		l = l - 1
		alreadyCut = true
	}
	if !alreadyCut && r < len(slice) && slice[r] == ' ' {
		r = r + 1
	}

	return append(slice[:l], slice[r:]...)
}

type BOCTxByDescription struct {
	Card        string
	Trace       string
	Auth        string
	Type        string
	Amount      string
	Date        time.Time
	Description string
	Country     string
}
