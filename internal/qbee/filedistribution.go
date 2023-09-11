package qbee

import "fmt"

const formtype_filedistribution = "file_distribution"

type FileDistributionService struct {
	Client *HttpClient
}

func (s FileDistributionService) Create(ct ConfigType, id string, filesets []FiledistributionFile, extend bool) (*Change, error) {
	p := ChangePayload{
		Formtype: formtype_filedistribution,
		Extend:   extend,
		Config: ChangePayloadConfig{
			Version: "v1",
			Enabled: true,
			Files:   filesets,
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

func (s FileDistributionService) Clear(ct ConfigType, id string) (*Change, error) {
	p := ChangePayload{
		Formtype: formtype_filedistribution,
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

func (s FileDistributionService) Get(ct ConfigType, id string) (*FileDistribution, error) {
	resp, err := s.Client.Configuration.GetConfiguration(ct, id)
	if err != nil {
		return nil, err
	}

	return resp.Config.BundleData.FileDistribution, nil
}
