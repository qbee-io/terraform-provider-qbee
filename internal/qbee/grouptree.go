package qbee

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type GrouptreeService struct {
	Client *HttpClient
}

type GetGrouptreeResponse struct {
	Title     string   `json:"title"`
	NodeId    string   `json:"node_id"`
	Type      string   `json:"type"`
	Ancestors []string `json:"ancestors"`
	Tags      []string `json:"tags"`
}

func (s GrouptreeService) Get(id string) (*GetGrouptreeResponse, error) {
	resp, err := s.Client.Get("/grouptree/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("GrouptreeService.Get: %w", err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GrouptreeService.Get io.ReadAll: %w", err)
	}

	var l GetGrouptreeResponse
	err = json.Unmarshal(b, &l)
	if err != nil {
		log.Printf("could not parse json: %v\n", string(b))
		return nil, fmt.Errorf("GrouptreeService.Get Unmarshal: %w", err)
	}

	return &l, nil
}

type grouptreeModificationOptions struct {
	Changes []grouptreeChanges `json:"changes"`
}

type grouptreeChanges struct {
	Action string                    `json:"action"`
	Data   grouptreeModificationData `json:"data"`
}

type grouptreeModificationData struct {
	ParentId    string `json:"parent_id"`
	OldParentId string `json:"old_parent_id"`
	Title       string `json:"title"`
	NodeId      string `json:"node_id"`
	Position    int    `json:"position"`
	Type        string `json:"type"`
}

func (s GrouptreeService) Create(id string, ancestor string, title string) error {
	changes := []grouptreeChanges{
		{
			Action: "create", Data: grouptreeModificationData{
				ParentId: ancestor,
				Title:    title,
				NodeId:   id,
				Position: 0,
				Type:     "group",
			}},
	}

	err := s.putGrouptreeModification(changes)
	if err != nil {
		return fmt.Errorf("GrouptreeService.Create: %w", err)
	}

	return nil
}

func (s GrouptreeService) Delete(id string, ancestor string) error {
	changes := []grouptreeChanges{
		{
			Action: "delete", Data: grouptreeModificationData{
				ParentId: ancestor,
				NodeId:   id,
				Type:     "group",
			}},
	}

	err := s.putGrouptreeModification(changes)
	if err != nil {
		return fmt.Errorf("GrouptreeService.Delete: %w", err)
	}

	return nil
}

func (s GrouptreeService) Rename(id string, ancestor string, title string) error {
	changes := []grouptreeChanges{
		{
			Action: "rename", Data: grouptreeModificationData{
				ParentId: ancestor,
				NodeId:   id,
				Title:    title,
			}},
	}

	err := s.putGrouptreeModification(changes)
	if err != nil {
		return fmt.Errorf("GrouptreeService.Rename: %w", err)
	}

	return nil
}

func (s GrouptreeService) Move(id string, oldAncestor string, newAncestor string) error {
	changes := []grouptreeChanges{
		{
			Action: "move", Data: grouptreeModificationData{
				ParentId:    newAncestor,
				OldParentId: oldAncestor,
				NodeId:      id,
			}},
	}

	err := s.putGrouptreeModification(changes)
	if err != nil {
		return fmt.Errorf("GrouptreeService.Move: %w", err)
	}

	return nil
}

func (s GrouptreeService) SetTags(id string, tags []string) error {
	_, err := s.Client.Patch("/grouptree/"+id, struct {
		Tags []string `json:"tags"`
	}{
		Tags: tags,
	})

	return err
}

func (s GrouptreeService) putGrouptreeModification(changes []grouptreeChanges) error {
	options := grouptreeModificationOptions{Changes: changes}
	_, err := s.Client.Put("/grouptree", options)
	return err
}
