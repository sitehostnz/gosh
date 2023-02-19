package models

type (
	StackServer struct {
		ID                  string   `json:"id"`
		ClientID            string   `json:"client_id"`
		Name                string   `json:"name"`
		Label               string   `json:"label"`
		State               string   `json:"state"`
		PrimaryIp           string   `json:"primary_ip"`
		IpAddrProxy         string   `json:"ip_addr_proxy"`
		Owner               bool     `json:"owner"`
		ImagesUsed          []string `json:"images_used"`
		ImagesRemaining     int      `json:"images_remaining"`
		ContainersRemaining int      `json:"containers_remaining"`
	}
)
