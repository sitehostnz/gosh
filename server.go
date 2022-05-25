package gosh

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type ServersService service

// Server represents a Server in the SiteHost.
type Server struct {
	Name             string `json:"name"`
	Label            string `json:"label"`
	ClientID         int    `json:"client_id"`
	Created          string `json:"created"`
	Type             string `json:"type"`
	RAM              string `json:"ram"`
	Root             string `json:"root"`
	Disk             int    `json:"disk"`
	Cores            string `json:"cores"`
	Core             string `json:"core"`
	Arch             string `json:"arch"`
	Kernel           string `json:"kernel"`
	Initrd           string `json:"initrd"`
	Modules          string `json:"modules"`
	Distro           string `json:"distro"`
	Os               string `json:"os"`
	Rescue           string `json:"rescue"`
	Managed          string `json:"managed"`
	Locked           string `json:"locked"`
	LockedFrom       string `json:"locked_from"`
	LockedUntil      string `json:"locked_until"`
	LockedComment    string `json:"locked_comment"`
	State            string `json:"state"`
	MaintDate        string `json:"maint_date"`
	MaintDateEnd     string `json:"maint_date_end"`
	EmailLogs        string `json:"email_logs"`
	VncPort          string `json:"vnc_port"`
	VncScreen        string `json:"vnc_screen"`
	IPAddrLimit      string `json:"ip_addr_limit"`
	Notes            string `json:"notes"`
	Ips              []IP
	Interfaces       []string      `json:"interfaces"`
	GroupID          string        `json:"group_id"`
	Partitions       []Partition   `json:"partitions"`
	BackupTypes      []interface{} `json:"backup_types"`
	AvailableKernels []Kernel      `json:"available_kernels"`
	ProductCode      string        `json:"product_code"`
	ProductType      string        `json:"product_type"`
	ProductName      string        `json:"product_name"`
	Subscription     Subscription  `json:"subscription"`
	LocationName     string        `json:"location_name"`
	LocationCode     string        `json:"location_code"`
	Mirror           string        `json:"mirror"`
	LastJob          Job           `json:"last_job"`
	BackupsEnabled   bool          `json:"backups_enabled"`
	BackupsProduct   interface{}   `json:"backups_product"`
	LocationNode     string        `json:"location_node"`
}

// IP represents the ips attached to the Server.
type IP struct {
	ID         string `json:"id"`
	ServerID   string `json:"server_id"`
	Bridge     string `json:"bridge"`
	MacAddr    string `json:"mac_addr"`
	IPType     string `json:"ip_type"`
	Prefix     string `json:"prefix"`
	Primary    bool   `json:"primary"`
	AddrFamily int    `json:"addr_family"`
	NetworkID  string `json:"network_id"`
	Rdns       string `json:"rdns"`
	IPAddr     string `json:"ip_addr"`
	Network    string `json:"network"`
	Gateway    string `json:"gateway"`
	Netmask    string `json:"netmask"`
	Broadcast  string `json:"broadcast"`
}

// Partition represents the disk partitions attached to the Server.
type Partition struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Device         string `json:"device"`
	Mountpoint     string `json:"mountpoint"`
	Size           string `json:"size"`
	NewSize        string `json:"new_size"`
	Fstype         string `json:"fstype"`
	Drbd           string `json:"drbd"`
	Backup         string `json:"backup"`
	DiskTotal      string `json:"disk_total"`
	DiskUsed       string `json:"disk_used"`
	InodesTotal    string `json:"inodes_total"`
	InodesUsed     string `json:"inodes_used"`
	AlertThreshold string `json:"alert_threshold"`
	Type           string `json:"type"`
}

// Kernel represents the kernel available to the Server.
type Kernel struct {
	Default    bool   `json:"default"`
	Kernel     string `json:"kernel"`
	Initrd     string `json:"initrd"`
	Modules    string `json:"modules"`
	Hypervisor string `json:"hypervisor"`
}

// Subscription represents the subscription attached to the Server.
type Subscription struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Price string `json:"price"`
}

// ServersGetResponse represents a result of a get Server call.
type ServersGetResponse struct {
	Server  Server `json:"return"`
	Message string `json:"msg"`
	Status  bool   `json:"status"`
}

// ServerCreateResponse represents a request to create a Server.
type ServerCreateResponse struct {
	Return struct {
		JobID    string   `json:"job_id"`
		Name     string   `json:"name"`
		Password string   `json:"password"`
		Ips      []string `json:"ips"`
		ServerID string   `json:"server_id"`
	} `json:"return"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}

// ServerCreateRequest represents a request to create a Server.
type ServerCreateRequest struct {
	ClientID    string        `json:"client_id"`
	Label       string        `json:"label"`
	Location    string        `json:"location"`
	ProductCode string        `json:"product_code"`
	Image       string        `json:"image"`
	Params      ParamsOptions `json:"params"`
}

// ParamsOptions represents the additionals parameters in the request to create a Server.
type ParamsOptions struct {
	Name      string   `json:"name,omitempty"`
	IPv4      []string `json:"ipv4"`
	IPv6      []string `json:"ipv6,omitempty"`
	SSHKeys   []string `json:"ssh_keys,omitempty"`
	ContactID string   `json:"contact_id,omitempty"`
	Backup    string   `json:"backup,omitempty"`
	SendEmail string   `json:"send_email,omitempty"`
}

// ServerDeleteRequest represents a request to delete a Server.
type ServerDeleteRequest struct {
	ClientID string `json:"client_id"`
	Name     string `json:"name"`
}

// ServerUpgradeRequest represents a request to upgrade a Server.
type ServerUpgradeRequest struct {
	Name string `json:"name"`
	Plan string `json:"plan"`
}

// ServerUpdateRequest represents a request to update a Server.
type ServerUpdateRequest struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// ServerDeleteResponse represents a result of a delete Server call.
type ServerDeleteResponse struct {
	Return struct {
		JobID string `json:"job_id"`
	} `json:"return"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}

// ServerCommitResponse represents a result of a commit changes Server call.
type ServerCommitResponse struct {
	Return struct {
		JobID string `json:"job_id"`
	} `json:"return"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}

// encodeSSHKeys encode the list of SSH Keys.
func (s *ServerCreateRequest) encodeSSHKeys() string {
	var buf strings.Builder

	if len(s.Params.SSHKeys) > 0 {
		for _, key := range s.Params.SSHKeys {
			buf.WriteString(fmt.Sprintf("&params[ssh_keys][]=%s", url.QueryEscape(key)))
		}
	}

	return buf.String()
}

// Get gets the Server with the provided name.
func (s *ServersService) Get(ctx context.Context, name string) (*Server, error) {
	u := fmt.Sprintf("server/get_server.json?name=%v", name)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(ServersGetResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return &response.Server, nil
}

// Create creates a Server.
func (s *ServersService) Create(ctx context.Context, opts *ServerCreateRequest) (*ServerCreateResponse, error) {
	u := "server/provision.json"

	keys := []string{
		"client_id",
		"label",
		"location",
		"product_code",
		"image",
		"params[ipv4]",
		"params[ssh_keys][]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("label", opts.Label)
	values.Add("location", opts.Location)
	values.Add("product_code", opts.ProductCode)
	values.Add("image", opts.Image)
	values.Add("params[ipv4]", "auto")

	if len(opts.Params.SSHKeys) > 0 {
		for _, key := range opts.Params.SSHKeys {
			values.Add("params[ssh_keys][]", key)
		}
	}

	req, err := s.client.NewRequest("POST", u, Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(ServerCreateResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Delete deletes a Server with the provided name.
func (s *ServersService) Delete(ctx context.Context, serverName string) (*ServerDeleteResponse, error) {
	u := "server/delete.json"

	keys := []string{
		"client_id",
		"name",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", serverName)

	req, err := s.client.NewRequest("POST", u, Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(ServerDeleteResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Update updates a Server with the provided name.
func (s *ServersService) Update(ctx context.Context, opts *ServerUpdateRequest) error {
	u := "server/update.json"

	keys := []string{
		"client_id",
		"name",
		"updates[label]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", opts.Name)
	values.Add("updates[label]", opts.Label)

	req, err := s.client.NewRequest("POST", u, Encode(values, keys))
	if err != nil {
		return err
	}

	return s.client.Do(ctx, req, nil)
}

// Upgrade upgrades a Server.
func (s *ServersService) Upgrade(ctx context.Context, opts *ServerUpgradeRequest) error {
	u := "server/upgrade_plan.json"

	keys := []string{
		"client_id",
		"name",
		"plan",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", opts.Name)
	values.Add("plan", opts.Plan)

	req, err := s.client.NewRequest("POST", u, Encode(values, keys))
	if err != nil {
		return err
	}

	return s.client.Do(ctx, req, nil)
}

// CommitChanges commit changes for upgrade a Server with the provided name.
func (s *ServersService) CommitChanges(ctx context.Context, serverName string) (*ServerCommitResponse, error) {
	u := "server/commit_disk_changes.json"

	keys := []string{
		"client_id",
		"name",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", serverName)

	req, err := s.client.NewRequest("POST", u, Encode(values, keys))
	if err != nil {
		return nil, err
	}

	response := new(ServerCommitResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
