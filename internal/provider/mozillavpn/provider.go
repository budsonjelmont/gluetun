package mozillavpn

import (
	"net/http"
	"github.com/qdm12/gluetun/internal/models"
)

type Provider struct {
	storage models.Storage
	client  *http.Client
}

func New(storage models.Storage, client *http.Client) *Provider {
	return &Provider{
		storage: storage,
		client:  client,
	}
}

func (p *Provider) Name() string {
	return "mozillavpn"
}
