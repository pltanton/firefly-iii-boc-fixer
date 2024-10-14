package app

import (
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_parseBOCDescription(t *testing.T) {
	type args struct {
		description string
	}
	tests := []struct {
		name string
		args args
		want BOCTxByDescription
	}{
		{
			"real example 1",
			args{"Card 1**2345 2024-10-10 71.41 EUR Auth 318622 Trace 357830 PURCHASE LU WWW.ALIEXPRESS.COM"},
			BOCTxByDescription{
				Card:        "1***2345",
				Trace:       "357830",
				Auth:        "318622",
				Type:        "PURCHASE",
				Amount:      "71.41 EUR",
				Date:        time.Date(2024, 10, 10, 0, 0, 0, 0, time.UTC),
				Description: "WWW.ALIEXPRESS.COM",
				Country:     "LU",
			},
		},
		{
			"real example 2",
			args{"Wolt PURCHASE CY Card 1***2345 2024-10-02 36.69 EUR Auth 366638 Trace 772220"},
			BOCTxByDescription{
				Card:        "1***2345",
				Trace:       "772220",
				Auth:        "366638",
				Type:        "PURCHASE",
				Amount:      "36.69 EUR",
				Date:        time.Date(2024, 10, 02, 0, 0, 0, 0, time.UTC),
				Description: "Wolt",
				Country:     "CY",
			},
		},
		{
			"INWARD",
			args{"INWARD CY000000000000 by HELLO HOLDINGS LIMITED>NOT PROVIDED>Salary forJune, 2023"},
			BOCTxByDescription{
				Type:        "INWARD",
				Description: "CY000000000000 by HELLO HOLDINGS LIMITED>NOT PROVIDED>Salary forJune, 2023",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseBOCDescription(slog.Default(), tt.args.description); !reflect.DeepEqual(got, tt.want) {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}
