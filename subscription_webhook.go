package phajay

const (
	// SubscriptionStatusConnected indicates accepted subscription authorization.
	SubscriptionStatusConnected = "SUBSCRIPTION_CONNECTED"
	// SubscriptionStatusSuccess indicates a successful subscription debit.
	SubscriptionStatusSuccess = "SUBSCRIPTION_SUCCESS"
	// SubscriptionStatusCanceled indicates a canceled authorization.
	SubscriptionStatusCanceled = "CANCEL"
)

// SubscriptionWebhookCallback describes setup, debit, and bank-app cancellation
// events. PaymentTransactionID is supplied only for debit notifications. Other
// fields are shared by all three documented event payloads.
type SubscriptionWebhookCallback struct {
	Message              string  `json:"message"`
	Status               string  `json:"status"`
	TransactionID        string  `json:"transactionId"`
	AuthCode             string  `json:"authCode"`
	Time                 string  `json:"time"`
	PaymentTransactionID *string `json:"paymentTransactionId"`
}
