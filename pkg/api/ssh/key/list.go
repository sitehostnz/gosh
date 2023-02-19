package key

import (
	"context"
	"fmt"
	"github.com/sitehostnz/gosh/pkg/models"
)

func (s *Client) List(ctx context.Context) (*[]models.SSHKey, error) {
	u := fmt.Sprintf("ssh/key/list_all.json")
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(ListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	// if the request works, don't care about pagination right now...
	return response.Return.SSHKeys, err
}
