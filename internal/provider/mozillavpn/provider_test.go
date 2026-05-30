package mozillavpn_test

import (
	"testing"

	"github.com/qdm12/gluetun/internal/constants/providers"
	"github.com/qdm12/gluetun/internal/provider/mozillavpn"
	"github.com/stretchr/testify/assert"
)

func TestProvider_Name(t *testing.T) {
	provider := mozillavpn.New(nil, nil)
	assert.Equal(t, providers.MozillaVPN, provider.Name())
}
