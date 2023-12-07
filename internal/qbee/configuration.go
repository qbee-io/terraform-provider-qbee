package qbee

import (
	"encoding/json"
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
	NodeId   string              `json:"node_id,omitempty"`
	Tag      string              `json:"tag,omitempty"`
	Formtype string              `json:"formtype"`
	Config   ChangePayloadConfig `json:"config"`
	Extend   bool                `json:"extend"`
}

type ChangePayloadConfig struct {
	Version      string                   `json:"version"`
	Enabled      bool                     `json:"enabled"`
	ResetToGroup bool                     `json:"reset_to_group,omitempty"`
	Files        []FiledistributionFile   `json:"files,omitempty"`
	Items        []SoftwareManagementItem `json:"items,omitempty"`
	Tables       FirewallTables           `json:"tables,omitempty"`
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

type GetConfigurationConfig struct {
	Id            string                  `json:"id"`
	Type          string                  `json:"type"`
	CommitId      string                  `json:"commit_id"`
	CommitCreated int64                   `json:"commit_created"`
	Bundles       []string                `json:"bundles"`
	BundleData    ConfigurationBundleData `json:"bundle_data"`
}

type GetConfigurationResponse struct {
	Config GetConfigurationConfig `json:"config"`
	Status string                 `json:"status"`
}

type ConfigurationBundleData struct {
	FileDistribution   *BundleConfiguration `json:"file_distribution"`
	SoftwareManagement *BundleConfiguration `json:"software_management"`
	Firewall           *BundleConfiguration `json:"firewall"`
}

func (c *ConfigurationBundleData) UnmarshalJSON(data []byte) error {
	// Because of a qbee bug, if a Configuration has no bundles, it will return as an
	// empty array instead of the expected empty map. This means we need to handle this specifically
	if data[0] == '{' {
		// It looks like an object; We can actually try unmarshalling it into the
		// ConfigurationBundleData.
		var v struct {
			FileDistribution   *BundleConfiguration `json:"file_distribution"`
			SoftwareManagement *BundleConfiguration `json:"software_management"`
			Firewall           *BundleConfiguration `json:"firewall"`
		}
		err := json.Unmarshal(data, &v)
		if err != nil {
			return err
		}

		c.FileDistribution = v.FileDistribution
		c.SoftwareManagement = v.SoftwareManagement
		c.Firewall = v.Firewall
	}

	return nil
}

type BundleConfiguration struct {
	Enabled               bool                     `json:"enabled"`
	Extend                bool                     `json:"extend"`
	Version               string                   `json:"version"`
	BundleCommitId        string                   `json:"bundle_commit_id"`
	FiledistributionFiles []FiledistributionFile   `json:"files"`
	SoftwareItems         []SoftwareManagementItem `json:"items"`
	FirewallTables        FirewallTables           `json:"tables"`
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

type FirewallTables struct {
	Filter FirewallFilter `json:"filter"`
}

type FirewallFilter struct {
	Input FirewallConfig `json:"INPUT"`
}

type FirewallConfig struct {
	Policy string         `json:"policy"`
	Rules  []FirewallRule `json:"rules"`
}

type FirewallRule struct {
	Proto   string `json:"proto"`
	Target  string `json:"target"`
	SrcIp   string `json:"srcIp"`
	DstPort string `json:"dstPort"`
}
