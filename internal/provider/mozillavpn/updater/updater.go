package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"sort"

	"github.com/qdm12/gluetun/internal/constants/vpn"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/common"
)

const relayAPI = "https://api.mullvad.net/public/relays/wireguard/v2"

type Updater struct {
	client *http.Client
}

func New(client *http.Client) *Updater {
	return &Updater{client: client}
}

type relayList struct {
	Wireguard struct {
		Relays []struct {
			Hostname   string `json:"hostname"`
			City       string `json:"city"`
			Country    string `json:"country"`
			Ipv4AddrIn string `json:"ipv4_addr_in"`
			Ipv6AddrIn string `json:"ipv6_addr_in"`
			Pubkey     string `json:"public_key"`
			Ports      []int  `json:"ports"`
		} `json:"relays"`
	} `json:"wireguard"`
}

func (u *Updater) FetchServers(ctx context.Context, minServers int) (servers []models.Server, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, relayAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching relays: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching relays: HTTP status code not OK: %d %s",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var data relayList
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decoding relays: %w", err)
	}

	servers = make([]models.Server, 0, len(data.Wireguard.Relays))
	for _, relay := range data.Wireguard.Relays {
		ips, err := parseRelayIPs(relay.Ipv4AddrIn, relay.Ipv6AddrIn)
		if err != nil {
			return nil, fmt.Errorf("parsing relay IP addresses for hostname %q: %w", relay.Hostname, err)
		}
		if len(ips) == 0 {
			continue
		}

		server := models.Server{
			VPN:      vpn.Wireguard,
			Hostname: relay.Hostname,
			Country:  relay.Country,
			City:     relay.City,
			IPs:      ips,
			WgPubKey: relay.Pubkey,
		}
		servers = append(servers, server)
	}

	if len(servers) < minServers {
		return nil, fmt.Errorf("%w: %d and expected at least %d",
			common.ErrNotEnoughServers, len(servers), minServers)
	}

	sort.Sort(models.SortableServers(servers))

	return servers, nil
}

func parseRelayIPs(ipv4AddrIn, ipv6AddrIn string) (ips []netip.Addr, err error) {
	ipStrings := []string{ipv4AddrIn, ipv6AddrIn}
	ips = make([]netip.Addr, 0, len(ipStrings))
	for _, ipString := range ipStrings {
		if ipString == "" {
			continue
		}

		ip, err := netip.ParseAddr(ipString)
		if err != nil {
			return nil, fmt.Errorf("parsing IP address %q: %w", ipString, err)
		}

		ips = append(ips, ip)
	}

	return ips, nil
}
