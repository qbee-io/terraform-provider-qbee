package qbee

import "fmt"

const formtype_firewall = "firewall"

type FirewallService struct {
	Client *HttpClient
}

func (s FirewallService) Create(ct ConfigType, id string, tables FirewallTables, extend bool) (*Change, error) {
	p := ChangePayload{
		Formtype: formtype_firewall,
		Extend:   extend,
		Config: ChangePayloadConfig{
			Version: "v1",
			Enabled: true,
			Tables:  tables,
		},
	}

	if ct == ConfigForTag {
		p.Tag = id
	} else {
		p.NodeId = id
	}

	r, err := s.Client.Configuration.CreateChange(p)
	if err != nil {
		return nil, fmt.Errorf("CreateTagFirewall: %w", err)
	}

	return r, nil
}

func (s FirewallService) Clear(ct ConfigType, id string) (*Change, error) {
	p := ChangePayload{
		Formtype: formtype_firewall,
		Config: ChangePayloadConfig{
			ResetToGroup: true,
		},
	}

	if ct == ConfigForTag {
		p.Tag = id
	} else {
		p.NodeId = id
	}

	r, err := s.Client.Configuration.CreateChange(p)
	if err != nil {
		return nil, fmt.Errorf("Clear: %w", err)
	}

	return r, nil
}

func (s FirewallService) Get(ct ConfigType, id string) (*BundleConfiguration, error) {
	resp, err := s.Client.Configuration.GetConfiguration(ct, id)
	if err != nil {
		return nil, err
	}

	return resp.Config.BundleData.Firewall, nil
}
