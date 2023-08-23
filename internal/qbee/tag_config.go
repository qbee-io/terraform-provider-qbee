package qbee

import "fmt"

const formtype_filedistribution = "file_distribution"

type TagConfigService struct {
	Client *HttpClient
}

func (s TagConfigService) CreateFileDistribution(tag string, filesets []FilesetConfig, extend bool) (*ChangeResponse, error) {
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
		return nil, fmt.Errorf("CreateFileDistribution: %w", err)
	}

	return r, nil
}

func (s TagConfigService) ClearFileDistribution(tag string) (*ChangeResponse, error) {
	p := ChangePayload{
		Tag:      tag,
		Formtype: formtype_filedistribution,
		Config: ChangePayloadConfig{
			ResetToGroup: true,
		},
	}

	r, err := s.Client.Configuration.CreateChange(p)
	if err != nil {
		return nil, fmt.Errorf("ClearFileDistribution: %w", err)
	}

	return r, nil
}

func (s TagConfigService) GetFiledistribution(tag string) (*GetFileDistributionResponse, error) {
	resp, err := s.Client.Configuration.GetConfiguration("tag", tag)
	if err != nil {
		return nil, err
	}

	return resp.Config.BundleData.FileDistribution, nil
}
