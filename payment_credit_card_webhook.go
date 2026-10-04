package phajay

// CreditCardDetails contains the masked card information supplied by Phajay.
// All fields are optional and may be null in the gateway payload.
type CreditCardDetails struct {
	IssuerCode           *string `json:"issuerCode"`
	CardHolderName       *string `json:"cardHolderName"`
	CardNumber           *string `json:"cardNumber"`
	CardExpiry           *string `json:"cardExpiry"`
	AuthenticationStatus *string `json:"authenticationStatus"`
	ECIValue             *string `json:"eciValue"`
	XIDValue             *string `json:"xidValue"`
	CAVVValue            *string `json:"cavvValue"`
	EnrolledFlag         *string `json:"enrolledFlag"`
}

// CreditCardTransactionAmount describes the card transaction's currency and
// amount. Fields remain nil when not supplied.
type CreditCardTransactionAmount struct {
	AmountText    *string  `json:"amountText"`
	CurrencyCode  *string  `json:"currencyCode"`
	DecimalPlaces *int     `json:"decimalPlaces"`
	Amount        *float64 `json:"amount"`
}

// CreditCardWebhookCallback is the card payment callback sent to the configured
// webhook URL. The card documentation does not guarantee a mandatory subset,
// so fields are pointers. Decode only after verifying the webhook signature
// using the guide and key supplied by Phajay support.
type CreditCardWebhookCallback struct {
	Message           *string                      `json:"message"`
	TransactionID     *string                      `json:"transactionId"`
	LinkCode          *string                      `json:"linkCode"`
	PaymentMethod     *string                      `json:"paymentMethod"`
	Status            *string                      `json:"status"`
	UserID            *string                      `json:"userId"`
	Description       *string                      `json:"description"`
	Tag1              *string                      `json:"tag1"`
	Tag2              *string                      `json:"tag2"`
	Tag3              *string                      `json:"tag3"`
	Tag4              *string                      `json:"tag4"`
	Tag5              *string                      `json:"tag5"`
	Tag6              *string                      `json:"tag6"`
	SuccessURL        *string                      `json:"successURL"`
	TxnAmount         *float64                     `json:"txnAmount"`
	Currency          *string                      `json:"currency"`
	CardType          *string                      `json:"cardType"`
	CreditCardDetails *CreditCardDetails           `json:"creditCardDetails"`
	TransactionAmount *CreditCardTransactionAmount `json:"transactionAmount"`
}
