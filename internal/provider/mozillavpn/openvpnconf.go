package mozillavpn

import "fmt"

func (p *Provider) OpenVPNConfig(selection interface{}) (string, error) {
	return "", fmt.Errorf("OpenVPN is not supported for MozillaVPN")
}
