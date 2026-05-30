package mozillavpn_test

import (
	"testing"

	"github.com/qdm12/gluetun/internal/configuration/settings"
	"github.com/qdm12/gluetun/internal/constants/providers"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/mozillavpn"
	"github.com/stretchr/testify/assert"
)

func TestProvider_Name(t *testing.T) {
	t.Parallel()

	provider := mozillavpn.New(nil, nil)
	assert.Equal(t, providers.MozillaVPN, provider.Name())
}

func TestProvider_OpenVPNConfig(t *testing.T) {
	t.Parallel()

	provider := mozillavpn.New(nil, nil)

	assert.Panics(t, func() {
		provider.OpenVPNConfig(models.Connection{}, settings.OpenVPN{}, false)
	})
}
