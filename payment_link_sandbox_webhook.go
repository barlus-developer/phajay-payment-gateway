package phajay

// SandboxPaymentLinkWebhookCallback models the separate sandbox link callback.
// Its field names differ from production (amount instead of txnAmount, user
// instead of userId, and _id). The documentation does not guarantee a mandatory
// subset, so fields remain nil when absent. Description uses the documented
// wire spelling "descriptione".
type SandboxPaymentLinkWebhookCallback struct {
	Message             *string  `json:"message"`
	ID                  *string  `json:"_id"`
	OrderNo             *string  `json:"orderNo"`
	LinkCode            *string  `json:"linkCode"`
	Amount              *float64 `json:"amount"`
	Currency            *string  `json:"currency"`
	Description         *string  `json:"descriptione"`
	Status              *string  `json:"status"`
	User                *string  `json:"user"`
	Tag1                *string  `json:"tag1"`
	Tag2                *string  `json:"tag2"`
	Tag3                *string  `json:"tag3"`
	Tag4                *string  `json:"tag4"`
	Tag5                *string  `json:"tag5"`
	Tag6                *string  `json:"tag6"`
	IsMerchantIDPayment *bool    `json:"isMerchantIdPayment"`
	IsAffiliatePayment  *bool    `json:"isAffiliatePayment"`
	CreatedAt           *string  `json:"createdAt"`
	UpdatedAt           *string  `json:"updateAt"`
}
