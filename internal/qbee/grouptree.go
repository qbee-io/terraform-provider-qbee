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
	OldParentId string `json:"oldParentId"`
	Title       string `json:"title"`
	NodeId      string `json:"node_id"`
	Position    int    `json:"position"`
	Type        string `json:"type"`
}

type GrouptreeModificationResponse struct {
	Sha     string        `json:"sha"`
	UserId  string        `json:"user_id"`
	Changes string        `json:"changes"`
	Type    string        `json:"type"`
	Message string        `json:"message"`
	Created uint64        `json:"created"`
	Error   ErrorResponse `json:"error"`
}

func (s GrouptreeService) Create(id string, ancestor string, title string) (*GrouptreeModificationResponse, error) {
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

	l, err := s.putGrouptreeModification(changes)
	if err != nil {
		return nil, fmt.Errorf("GrouptreeService.Create: %w", err)
	}

	return l, nil
}

func (s GrouptreeService) Delete(id string, ancestor string) (*GrouptreeModificationResponse, error) {
	changes := []grouptreeChanges{
		{
			Action: "delete", Data: grouptreeModificationData{
				ParentId: ancestor,
				NodeId:   id,
				Type:     "group",
			}},
	}

	resp, err := s.putGrouptreeModification(changes)
	if err != nil {
		return nil, fmt.Errorf("GrouptreeService.Delete: %w", err)
	}

	return resp, nil
}

func (s GrouptreeService) Rename(id string, ancestor string, title string) (*GrouptreeModificationResponse, error) {
	changes := []grouptreeChanges{
		{
			Action: "rename", Data: grouptreeModificationData{
				ParentId: ancestor,
				NodeId:   id,
				Title:    title,
			}},
	}

	resp, err := s.putGrouptreeModification(changes)
	if err != nil {
		return nil, fmt.Errorf("GrouptreeService.Delete: %w", err)
	}

	return resp, nil
}

func (s GrouptreeService) Move(id string, oldAncestor string, newAncestor string) (*GrouptreeModificationResponse, error) {
	changes := []grouptreeChanges{
		{
			Action: "move", Data: grouptreeModificationData{
				ParentId:    newAncestor,
				OldParentId: oldAncestor,
				NodeId:      id,
			}},
	}

	resp, err := s.putGrouptreeModification(changes)
	if err != nil {
		return nil, fmt.Errorf("GrouptreeService.Delete: %w", err)
	}

	return resp, nil
}

func (s GrouptreeService) putGrouptreeModification(changes []grouptreeChanges) (*GrouptreeModificationResponse, error) {
	options := grouptreeModificationOptions{Changes: changes}
	resp, err := s.Client.Put("/grouptree", options)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	var l GrouptreeModificationResponse
	err = json.Unmarshal(b, &l)
	if err != nil {
		log.Printf("could not parse json: %v\n", string(b))
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &l, nil
}