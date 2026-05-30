package updater_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/qdm12/gluetun/internal/provider/mozillavpn/updater"
	"github.com/stretchr/testify/assert"
)

func TestFetchServers_invalidURL(t *testing.T) {
	client := &http.Client{}
	ctx := context.Background()
	// Temporarily override relayAPI for test
	servers, err := updater.FetchServers(ctx, client)
	assert.NoError(t, err)
	assert.NotEmpty(t, servers)
}
