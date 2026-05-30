package updater_test

import (
	"context"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"testing"

	"github.com/qdm12/gluetun/internal/constants/vpn"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/common"
	"github.com/qdm12/gluetun/internal/provider/mozillavpn/updater"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestUpdater_FetchServers(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		minServers     int
		responseStatus int
		responseBody   string
		errMessage     string
		errIs          error
		servers        []models.Server
	}{
		"response_status_not_ok": {
			responseStatus: http.StatusNoContent,
			errMessage:     "fetching relays: HTTP status code not OK: 204 No Content",
		},
		"decode_error": {
			responseStatus: http.StatusOK,
			responseBody:   "{",
			errMessage:     "decoding relays: unexpected EOF",
		},
		"invalid_ip": {
			responseStatus: http.StatusOK,
			responseBody: `{
				"wireguard": {
					"relays": [
						{
							"hostname": "relay1",
							"country": "Country1",
							"city": "City1",
							"ipv4_addr_in": "not_an_ip",
							"pubkey": "publickey1"
						}
					]
				}
			}`,
			errMessage: "parsing relay IP addresses for hostname \"relay1\": parsing IP address \"not_an_ip\"",
		},
		"not_enough_servers": {
			minServers:     2,
			responseStatus: http.StatusOK,
			responseBody: `{
				"wireguard": {
					"relays": [
						{
							"hostname": "relay1",
							"country": "Country1",
							"city": "City1",
							"ipv4_addr_in": "1.2.3.4",
							"pubkey": "publickey1"
						}
					]
				}
			}`,
			errMessage: "not enough servers found: 1 and expected at least 2",
			errIs:      common.ErrNotEnoughServers,
		},
		"success": {
			minServers:     1,
			responseStatus: http.StatusOK,
			responseBody: `{
				"wireguard": {
					"relays": [
						{
							"hostname": "relay1",
							"country": "Country1",
							"city": "City1",
							"ipv4_addr_in": "1.2.3.4",
							"ipv6_addr_in": "2001:db8::1",
							"pubkey": "publickey1"
						}
					]
				}
			}`,
			servers: []models.Server{{
				VPN:      vpn.Wireguard,
				Country:  "Country1",
				City:     "City1",
				Hostname: "relay1",
				WgPubKey: "publickey1",
				IPs: []netip.Addr{
					netip.AddrFrom4([4]byte{1, 2, 3, 4}),
					netip.MustParseAddr("2001:db8::1"),
				},
			}},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			client := &http.Client{
				Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, "https://api.mullvad.net/public/relays/wireguard/v2", r.URL.String())
					return &http.Response{
						StatusCode: testCase.responseStatus,
						Status:     http.StatusText(testCase.responseStatus),
						Body:       io.NopCloser(strings.NewReader(testCase.responseBody)),
					}, nil
				}),
			}

			updater := updater.New(client)

			servers, err := updater.FetchServers(ctx, testCase.minServers)

			assert.Equal(t, testCase.servers, servers)
			if testCase.errMessage != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, testCase.errMessage)
				if testCase.errIs != nil {
					assert.ErrorIs(t, err, testCase.errIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
