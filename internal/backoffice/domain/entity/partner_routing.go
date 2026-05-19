package entity

type PartnerRoutingProfile struct {
	ID                    string   `json:"id"`
	Code                  string   `json:"code"`
	Name                  string   `json:"name"`
	PartnerType           string   `json:"partnerType"`
	Status                string   `json:"status"`
	SupportedProductTypes []string `json:"supportedProductTypes"`
	SupportedRegions      []string `json:"supportedRegions"`
	SLADays               int32    `json:"slaDays"`
	RoutingPriority       int32    `json:"routingPriority"`
}

type RoutingPartnerOption struct {
	Partner  PartnerRoutingProfile `json:"partner"`
	Eligible bool                  `json:"eligible"`
	Reason   string                `json:"reason"`
}

type RoutedOrderRecommendation struct {
	CandidateID      string                 `json:"candidateId"`
	ProductTitle     string                 `json:"productTitle"`
	CandidatePartner string                 `json:"candidatePartner"`
	ProductType      string                 `json:"productType"`
	ShipRegion       string                 `json:"shipRegion"`
	SelectedPartner  string                 `json:"selectedPartner"`
	Summary          string                 `json:"summary"`
	Options          []RoutingPartnerOption `json:"options"`
}
