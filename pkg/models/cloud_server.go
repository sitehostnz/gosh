package models

type (
	// CloudServer is a view of a server that can run stacks on. This server also exists under the server API.
	CloudServer struct {
		ID                  string   `json:"id"`
		ClientID            string   `json:"client_id"`
		Name                string   `json:"name"`
		Label               string   `json:"label"`
		State               string   `json:"state"`
		PrimaryIP           string   `json:"primary_ip"`
		IPAddrProxy         string   `json:"ip_addr_proxy"`
		ProductID           string   `json:"product_id"`
		Owner               bool     `json:"owner"`
		ImagesUsed          []string `json:"images_used"`
		ImagesRemaining     int      `json:"images_remaining"`
		ContainersRemaining int      `json:"containers_remaining"`
	}
)
