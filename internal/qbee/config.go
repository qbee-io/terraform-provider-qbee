package qbee

import "fmt"

const formtype_filedistribution = "file_distribution"

type ConfigService struct {
	Client *HttpClient
}

func (s ConfigService) CreateTagFileDistribution(tag string, filesets []FilesetConfig, extend bool) (*ChangeResponse, error) {
	p := ChangePayload{
		Tag:      tag,
		Formtype: formtype_filedistribution,
		Extend:   extend,
		Config: ChangePayloadConfig{
			Version: "v1",
			Enabled: false,
			Files:   filesets,
		},
	}

	r, err := s.Client.Configuration.CreateChange(p)
	if err != nil {
		return nil, fmt.Errorf("CreateTagFileDistribution: %w", err)
	}

	return r, nil
}

func (s ConfigService) CreateNodeFileDistribution(nodeId string, filesets []FilesetConfig, extend bool) (*ChangeResponse, error) {
	p := ChangePayload{
		NodeId:   nodeId,
		Formtype: formtype_filedistribution,
		Extend:   extend,
		Config: ChangePayloadConfig{
			Version: "v1",
			Enabled: false,
			Files:   filesets,
		},
	}

	r, err := s.Client.Configuration.CreateChange(p)
	if err != nil {
		return nil, fmt.Errorf("CreateNodeFileDistribution: %w", err)
	}

	return r, nil
}

func (s ConfigService) ClearTagFileDistribution(tag string) (*ChangeResponse, error) {
	p := ChangePayload{
		Tag:      tag,
		Formtype: formtype_filedistribution,
		Config: ChangePayloadConfig{
			ResetToGroup: true,
		},
	}

	r, err := s.Client.Configuration.CreateChange(p)
	if err != nil {
		return nil, fmt.Errorf("ClearTagFileDistribution: %w", err)
	}

	return r, nil
}

func (s ConfigService) ClearNodeFileDistribution(nodeId string) (*ChangeResponse, error) {
	p := ChangePayload{
		NodeId:   nodeId,
		Formtype: formtype_filedistribution,
		Config: ChangePayloadConfig{
			ResetToGroup: true,
		},
	}

	r, err := s.Client.Configuration.CreateChange(p)
	if err != nil {
		return nil, fmt.Errorf("ClearNodeFileDistribution: %w", err)
	}

	return r, nil
}

func (s ConfigService) GetTagFiledistribution(tag string) (*GetFileDistributionResponse, error) {
	resp, err := s.Client.Configuration.GetConfiguration("tag", tag)
	if err != nil {
		return nil, err
	}

	return resp.Config.BundleData.FileDistribution, nil
}

func (s ConfigService) GetNodeFiledistribution(tag string) (*GetFileDistributionResponse, error) {
	resp, err := s.Client.Configuration.GetConfiguration("node", tag)
	if err != nil {
		return nil, err
	}

	return resp.Config.BundleData.FileDistribution, nil
}
