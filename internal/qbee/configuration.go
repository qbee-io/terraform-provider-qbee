package qbee

import (
	"fmt"
)

type ConfigurationService struct {
	Client *HttpClient
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
	Labels  []string `json:"labels"`
	Message string   `json:"message"`
	Created int      `json:"created"`
	// Only this is set if no changes
	Result string `json:"result"`
}

func (r CommitResponse) HadChanges() bool {
	return r.Result != ""
}

type FilesetParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FilesetTemplate struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	IsTemplate  bool   `json:"is_template"`
}

type FilesetConfig struct {
	PreCondition string             `json:"pre_condition"`
	Command      string             `json:"command"`
	Templates    []FilesetTemplate  `json:"templates"`
	Parameters   []FilesetParameter `json:"parameters"`
}

type ChangePayload struct {
	NodeId   string              `json:"node_id"`
	Tag      string              `json:"tag"`
	Formtype string              `json:"formtype"`
	Config   ChangePayloadConfig `json:"config"`
	Extend   bool                `json:"extend"`
}

type ChangePayloadConfig struct {
	Version      string          `json:"version"`
	Enabled      bool            `json:"enabled"`
	ResetToGroup bool            `json:"reset_to_group"`
	Files        []FilesetConfig `json:"files"`
}

type ChangeResponse struct {
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
			BundleCommitId string        `json:"bundle_commit_id"`
		} `json:"config"`
	} `json:"content"`
}

func (s ConfigurationService) CreateChange(change ChangePayload) (*ChangeResponse, error) {
	r, err := s.Client.Post("/change", change)
	if err != nil {
		return nil, err
	}

	var response ChangeResponse

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

func (s ConfigurationService) DeleteUncommitted(sha string) error {
	_, err := s.Client.Delete("/change/"+sha, nil)
	if err != nil {
		return err
	}

	return nil
}

type GetConfigurationResponse struct {
	Config struct {
		Id            string   `json:"id"`
		Type          string   `json:"type"`
		CommitId      string   `json:"commit_id"`
		CommitCreated int64    `json:"commit_created"`
		Bundles       []string `json:"bundles"`
		BundleData    struct {
			FileDistribution *GetFileDistributionResponse `json:"file_distribution"`
		} `json:"bundle_data"`
	} `json:"config"`
	Status string `json:"status"`
}

type GetFileDistributionResponse struct {
	Enabled        bool                           `json:"enabled"`
	Extend         bool                           `json:"extend"`
	Version        string                         `json:"version"`
	BundleCommitId string                         `json:"bundle_commit_id"`
	Files          []FileDistributionFileResponse `json:"files"`
}

type FileDistributionFileResponse struct {
	Command      string                                  `json:"command"`
	PreCondition string                                  `json:"pre_condition"`
	Templates    []FiledistributionFileTemplateResponse  `json:"templates"`
	Parameters   []FiledistributionFileParameterResponse `json:"parameters"`
}

type FiledistributionFileTemplateResponse struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	IsTemplate  bool   `json:"is_template"`
}
type FiledistributionFileParameterResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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
