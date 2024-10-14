package app

type WebhookRequest struct {
	Content struct {
		ID           int64         `json:"id"`
		Transactions []Transaction `json:"transactions"`
	} `json:"content"`
}

type Transaction struct {
	SourceIBAN      string   `json:"source_iban"`
	DestinationIBAN string   `json:"destination_iban"`
	Description     string   `json:"description"`
	Type            string   `json:"type"`
	Tags            []string `json:"tags"`
}
