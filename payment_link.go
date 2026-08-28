package phajay

// ---- PaymentLink Struct ----

type PaymentLink struct {
	Request  paymentLinkRequest
	Response paymentLinkResponse
}

type paymentLinkRequest struct {
	OrderNo     string `json:"orderNo"`
	Amount      int    `json:"amount"`      // required
	Description string `json:"description"` // required
	Tag1        string `json:"tag1"`
	Tag2        string `json:"tag2"`
	Tag3        string `json:"tag3"`
}

type paymentLinkResponse struct {
	Message     string `json:"message"`
	RedirectURL string `json:"redirectURL"`
	OrderNo     string `json:"orderNo"`
}

// ---- PaymentLink Usecase ----

func (p *Phajay) CreatePaymentLink(request paymentLinkRequest) (paymentLinkResponse, error) {
	return paymentLinkResponse{}, nil
}