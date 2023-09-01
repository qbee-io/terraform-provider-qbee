package qbee

import "fmt"

const formtype_filedistribution = "file_distribution"

type ConfigService struct {
	Client *HttpClient
}

type ConfigType int

const (
	ConfigForTag ConfigType = iota
	ConfigForNode
)

func (t ConfigType) String() string {
	return []string{"tag", "node"}[t]
}

func (s ConfigService) CreateFileDistribution(ct ConfigType, id string, filesets []FilesetConfig, extend bool) (*ChangeResponse, error) {
	p := ChangePayload{
		Formtype: formtype_filedistribution,
		Extend:   extend,
		Config: ChangePayloadConfig{
			Version: "v1",
			Enabled: false,
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

// Deprecated: use CreateFiledistribution with a ConfigType instead
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

// Deprecated: use CreateFiledistribution with a ConfigType instead
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

func (s ConfigService) ClearFileDistribution(ct ConfigType, id string) (*ChangeResponse, error) {
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
		return nil, fmt.Errorf("ClearFileDistribution: %w", err)
	}

	return r, nil
}

// Deprecated: use ClearFiledistribution with a ConfigType instead
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

// Deprecated: use ClearFiledistribution with a ConfigType instead
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

func (s ConfigService) GetFiledistribution(ct ConfigType, id string) (*GetFileDistributionResponse, error) {
	resp, err := s.Client.Configuration.GetConfiguration(ct, id)
	if err != nil {
		return nil, err
	}

	return resp.Config.BundleData.FileDistribution, nil
}

// Deprecated: use GetFiledistribution with a ConfigType instead
func (s ConfigService) GetTagFiledistribution(tag string) (*GetFileDistributionResponse, error) {
	resp, err := s.Client.Configuration.GetConfiguration(ConfigForTag, tag)
	if err != nil {
		return nil, err
	}

	return resp.Config.BundleData.FileDistribution, nil
}

// Deprecated: use GetFiledistribution with a ConfigType instead
func (s ConfigService) GetNodeFiledistribution(tag string) (*GetFileDistributionResponse, error) {
	resp, err := s.Client.Configuration.GetConfiguration(ConfigForNode, tag)
	if err != nil {
		return nil, err
	}

	return resp.Config.BundleData.FileDistribution, nil
}
