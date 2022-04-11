package gosh

import (
	"context"
	"fmt"
)

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
	Ips              []IPs
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

type IPs struct {
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

type Kernel struct {
	Default    bool   `json:"default"`
	Kernel     string `json:"kernel"`
	Initrd     string `json:"initrd"`
	Modules    string `json:"modules"`
	Hypervisor string `json:"hypervisor"`
}

type Subscription struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Price string `json:"price"`
}

type Job struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	State string `json:"state"`
}

type ServersResponse struct {
	Server  Server `json:"return"`
	Message string `json:"msg"`
	Status  bool   `json:"status"`
}

// GetServer gets the server information with the provided Name
func (c *Client) GetServer(ctx context.Context, serverName string) (*Server, error) {
	e, err := c.Servers.Endpoint()
	if err != nil {
		return nil, err
	}

	serverResponse := &ServersResponse{}
	req := c.R(ctx)
	req.SetQueryParam("name", serverName)

	e = fmt.Sprintf("%s/get_server.json", e)

	_, err = coupleAPIErrors(req.SetResult(serverResponse).Get(e))
	if err != nil {
		return nil, err
	}

	return &serverResponse.Server, nil
}

// DeleteServer delete a server with the provided Name
func (c *Client) DeleteServer(ctx context.Context, serverName string) error {
	e, err := c.Servers.Endpoint()
	if err != nil {
		return err
	}

	req := c.R(ctx)
	req.SetQueryParam("name", serverName)

	e = fmt.Sprintf("%s/delete.json", e)

	_, err = coupleAPIErrors(req.Delete(e))

	return err
}
