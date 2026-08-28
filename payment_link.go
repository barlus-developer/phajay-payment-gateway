package phajay

type PaymentLink struct {
	Request paymentLinkRequest
}

type paymentLinkRequest struct {
	OrderNo     string `json:"orderNo"`
	Amount      int    `json:"amount"` // required
	Description string `json:"description"` // required
	Tag1        string `json:"tag1"`
	Tag2        string `json:"tag2"`
	Tag3        string `json:"tag3"`
}
