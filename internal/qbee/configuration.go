package qbee

import (
	"fmt"
)

type ConfigurationService struct {
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

func (s ConfigurationService) CreateChange(change ChangePayload) (*Change, error) {
	r, err := s.Client.Post("/change", change)
	if err != nil {
		return nil, err
	}

	var response Change

	err = s.Client.ParseJsonBody(r, &response)
	if err != nil {
		return nil, fmt.Errorf("ConfigurationService.CreateChange(%+v): %w", change, err)
	}

	return &response, nil
}

func (s ConfigurationService) Commit(commitMessage string) (*CommitResponse, error) {
	r, err := s.Client.Post("/commit", CommitPayload{
		Action:  "commit",
		Message: commitMessage,
	})
	if err != nil {
		return nil, err
	}

	var response CommitResponse

	err = s.Client.ParseJsonBody(r, &response)
	if err != nil {
		return nil, fmt.Errorf("ConfigurationService.Commit(%v): %w", commitMessage, err)
	}

	return &response, nil
}

func (s ConfigurationService) ClearUncommitted() error {
	_, err := s.Client.Delete("/changes", nil)
	return err
}

func (s ConfigurationService) GetUncommitted() ([]Change, error) {
	r, err := s.Client.Get("/changelist", nil)
	if err != nil {
		return nil, err
	}

	var response []Change
	err = s.Client.ParseJsonBody(r, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s ConfigurationService) DeleteUncommitted(sha string) error {
	_, err := s.Client.Delete("/change/"+sha, nil)
	return err
}
func (s ConfigurationService) GetConfiguration(ct ConfigType, identifier string) (*GetConfigurationResponse, error) {
	r, err := s.Client.Get(fmt.Sprintf("/config/%v/%v", ct.String(), identifier), nil)
	if err != nil {
		return nil, fmt.Errorf("ConfigurationService.GetConfiguration(%v, %v): %w", ct.String(), identifier, err)
	}

	var response GetConfigurationResponse

	err = s.Client.ParseJsonBody(r, &response)
	if err != nil {
		return nil, fmt.Errorf("ConfigurationService.GetConfiguration(%v, %v): %w", ct.String(), identifier, err)
	}

	return &response, nil
}

type CommitPayload struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

type CommitResponse struct {
	Sha     string   `json:"sha"`
	UserId  string   `json:"user_id"`
	Changes []string `json:"changes"`
	Type    string   `json:"type"`
	Message string   `json:"message"`
	Created int      `json:"created"`
	// Only this is set if no changes
	Result string `json:"result"`
}

type ChangePayload struct {
	NodeId   string              `json:"node_id"`
	Tag      string              `json:"tag"`
	Formtype string              `json:"formtype"`
	Config   ChangePayloadConfig `json:"config"`
	Extend   bool                `json:"extend"`
}

type ChangePayloadConfig struct {
	Version      string                   `json:"version"`
	Enabled      bool                     `json:"enabled"`
	ResetToGroup bool                     `json:"reset_to_group"`
	Files        []FiledistributionFile   `json:"files"`
	Items        []SoftwareManagementItem `json:"items"`
}

type Change struct {
	Id      string `json:"id"`
	UserId  string `json:"user_id"`
	Type    string `json:"type"`
	Created int    `json:"created"`
	Status  string `json:"status"`
	Sha     string `json:"sha"`
	Content struct {
		Tag      string `json:"tag"`
		NodeId   string `json:"node_id"`
		FormType string `json:"form_type"`
		Config   struct {
			Enabled        bool          `json:"enabled"`
			Version        string        `json:"version"`
			Files          []interface{} `json:"files"`
			Items          []interface{} `json:"items"`
			BundleCommitId string        `json:"bundle_commit_id"`
		} `json:"config"`
	} `json:"content"`
}

type GetConfigurationResponse struct {
	Config struct {
		Id            string   `json:"id"`
		Type          string   `json:"type"`
		CommitId      string   `json:"commit_id"`
		CommitCreated int64    `json:"commit_created"`
		Bundles       []string `json:"bundles"`
		BundleData    struct {
			FileDistribution   *FileDistribution   `json:"file_distribution"`
			SoftwareManagement *SoftwareManagement `json:"software_management"`
		} `json:"bundle_data"`
	} `json:"config"`
	Status string `json:"status"`
}

type FileDistribution struct {
	Enabled        bool                   `json:"enabled"`
	Extend         bool                   `json:"extend"`
	Version        string                 `json:"version"`
	BundleCommitId string                 `json:"bundle_commit_id"`
	Files          []FiledistributionFile `json:"files"`
}

type FiledistributionFile struct {
	Command      string                      `json:"command"`
	PreCondition string                      `json:"pre_condition"`
	Templates    []FiledistributionTemplate  `json:"templates"`
	Parameters   []FiledistributionParameter `json:"parameters"`
}

type FiledistributionTemplate struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	IsTemplate  bool   `json:"is_template"`
}
type FiledistributionParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SoftwareManagement struct {
	Enabled        bool                     `json:"enabled"`
	Extend         bool                     `json:"extend"`
	Version        string                   `json:"version"`
	BundleCommitId string                   `json:"bundle_commit_id"`
	Items          []SoftwareManagementItem `json:"items"`
}

type SoftwareManagementItem struct {
	Package      string                         `json:"package"`
	ServiceName  string                         `json:"service_name"`
	PreCondition string                         `json:"pre_condition"`
	ConfigFiles  []SoftwareManagementConfigFile `json:"config_files"`
	Parameters   []SoftwareManagementParameter  `json:"parameters"`
}

type SoftwareManagementConfigFile struct {
	ConfigTemplate string `json:"config_template"`
	ConfigLocation string `json:"config_location"`
}

type SoftwareManagementParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
