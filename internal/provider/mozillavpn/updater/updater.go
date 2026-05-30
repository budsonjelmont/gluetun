package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/qdm12/gluetun/internal/models"
)

const relayAPI = "https://api.mullvad.net/public/relays/wireguard/v2"

type relayList struct {
	Wireguard struct {
		Relays []struct {
			Hostname   string   `json:"hostname"`
			City       string   `json:"city"`
			Country    string   `json:"country"`
			Ipv4AddrIn string   `json:"ipv4_addr_in"`
			Ipv6AddrIn string   `json:"ipv6_addr_in"`
			Pubkey     string   `json:"pubkey"`
			Ports      []int    `json:"ports"`
		} `json:"relays"`
	} `json:"wireguard"`
}

func FetchServers(ctx context.Context, client *http.Client) ([]models.Server, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, relayAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching relays: %w", err)
	}
	defer resp.Body.Close()
	var data relayList
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decoding relays: %w", err)
	}
	var servers []models.Server
	for _, relay := range data.Wireguard.Relays {
		server := models.Server{
			VPN:      "wireguard",
			Hostname: relay.Hostname,
			Country:  relay.Country,
			City:     relay.City,
			IPs:      []string{relay.Ipv4AddrIn, relay.Ipv6AddrIn},
			WgPubKey: relay.Pubkey,
		}
		servers = append(servers, server)
	}
	return servers, nil
}
