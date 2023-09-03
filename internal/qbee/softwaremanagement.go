package qbee

import "fmt"

const formtype_softwaremanagement = "software_management"

type SoftwaremanagementService struct {
	Client *HttpClient
}

func (s SoftwaremanagementService) Create(ct ConfigType, id string, items []SoftwareManagementItem, extend bool) (*Change, error) {
	p := ChangePayload{
		Formtype: formtype_softwaremanagement,
		Extend:   extend,
		Config: ChangePayloadConfig{
			Version: "v1",
			Enabled: false,
			Items:   items,
		},
	}

	if ct == ConfigForTag {
		p.Tag = id
	} else {
		p.NodeId = id
	}

	r, err := s.Client.Configuration.CreateChange(p)
	if err != nil {
		return nil, fmt.Errorf("CreateTagFileDistribution: %w", err)
	}

	return r, nil
}

func (s SoftwaremanagementService) Clear(ct ConfigType, id string) (*Change, error) {
	p := ChangePayload{
		Formtype: formtype_softwaremanagement,
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

func (s SoftwaremanagementService) Get(ct ConfigType, id string) (*SoftwareManagement, error) {
	resp, err := s.Client.Configuration.GetConfiguration(ct, id)
	if err != nil {
		return nil, err
	}

	return resp.Config.BundleData.SoftwareManagement, nil
}
