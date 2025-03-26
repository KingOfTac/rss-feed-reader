package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var endpoint = "https://europe-west1-tagr-prod.cloudfunctions.net/addToNodeV2"
var schemaNodeId = "SCHEMA"

const coreTemplateId = "SYS_T01"
const attrDefTemplateId = "SYS_T02"

type TanaAPIHelper struct {
	Token    string
	Endpoint string
}

type APIPlainNode struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Supertags   []SuperTag `json:"supertags,omitempty"`
}

type APINode struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Supertags   []SuperTag `json:"supertags,omitempty"`
}

type APIField struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SuperTag struct {
	ID string `json:"id"`
}

type TanaNode struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	NodeID      string `json:"nodeId"`
}

func NewTanaAPIHelper(token, endpointUrl string) *TanaAPIHelper {
	helper := &TanaAPIHelper{
		Token:    token,
		Endpoint: endpoint,
	}

	if endpointUrl != "" {
		helper.Endpoint = endpointUrl
	}

	return helper
}

func (h *TanaAPIHelper) createFieldDefinitions(fields []APIPlainNode) ([]TanaNode, error) {
	for i := range fields {
		fields[i].Supertags = append(fields[i].Supertags, SuperTag{ID: attrDefTemplateId})
	}

	payload := map[string]interface{}{
		"targetNodeId": schemaNodeId,
		"nodes":        fields,
	}

	createdFields, err := h.makeRequest(payload)
	if err != nil {
		return nil, err
	}

	var result []TanaNode
	for _, field := range createdFields {
		result = append(result, TanaNode{
			Name:        field.Name,
			Description: field.Description,
			NodeID:      field.NodeID,
		})
	}

	return result, nil
}

func (h *TanaAPIHelper) createTagDefinition(node APIPlainNode) (string, error) {
	node.Supertags = append(node.Supertags, SuperTag{ID: coreTemplateId})
	payload := map[string]interface{}{
		"targetNodeId": schemaNodeId,
		"nodes":        []APIPlainNode{node},
	}

	createdTag, err := h.makeRequest(payload)
	if err != nil {
		return "", err
	}

	return createdTag[0].NodeID, nil
}

func (h *TanaAPIHelper) createNode(node APINode, targetNodeId string) (TanaNode, error) {
	payload := map[string]interface{}{
		"targetNodeId": targetNodeId,
		"nodes":        []APINode{node},
	}

	createdNode, err := h.makeRequest(payload)
	if err != nil {
		return TanaNode{}, err
	}

	return createdNode[0], nil
}

func (h *TanaAPIHelper) setNodeName(newName, targetNodeId string) (TanaNode, error) {
	payload := map[string]interface{}{
		"targetNodeId": targetNodeId,
		"setName":      newName,
	}

	createdNode, err := h.makeRequest(payload)
	if err != nil {
		return TanaNode{}, err
	}

	return createdNode[0], nil
}

func (h *TanaAPIHelper) addField(field APIField, targetNodeId string) (TanaNode, error) {
	payload := map[string]interface{}{
		"targetNodeId": targetNodeId,
		"nodes":        []APIField{field},
	}

	createdNode, err := h.makeRequest(payload)
	if err != nil {
		return TanaNode{}, err
	}

	return createdNode[0], nil
}

func (h *TanaAPIHelper) makeRequest(payload interface{}) ([]TanaNode, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", h.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+h.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		var result struct {
			Children []TanaNode `json:"children"`
		}

		bodyBytes, _ := io.ReadAll(res.Body)
		err = json.Unmarshal(bodyBytes, &result)
		if err != nil {
			return nil, err
		}

		return result.Children, nil
	}

	bodyBytes, _ := io.ReadAll(res.Body)
	return nil, fmt.Errorf("error: %s", string(bodyBytes))
}
