package models

type UpdateUserSettings struct {
	SlackId    string   `json:"slackId"`
	StartTime  []string `json:"startTime"`
	EndTime    []string `json:"endTime"`
	AddressIds []string `json:"addressIds"`
}

type OrderUpdate struct {
	Order *ZomatoOrder `json:"order"`
}

type ZomatoOrder struct {
	OrderId         uint64           `json:"orderId"`
	Status          int              `json:"status"`
	PaymentStatus   int              `json:"paymentStatus"`
	DeliveryDetails *DeliveryDetails `json:"deliveryDetails"`
	ResInfo         *ResInfo         `json:"resInfo"`
}

type DeliveryDetails struct {
	DeliveryStatus  int    `json:"deliveryStatus"`
	DeliveryLabel   string `json:"deliveryLabel"`
	DeliveryMessage string `json:"deliveryMessage"`
}

type ResInfo struct {
	Name string `json:"name"`
}
