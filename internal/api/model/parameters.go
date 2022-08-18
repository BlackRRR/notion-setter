package model

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	paramsPath     = "assets/params"
	jsonFormatName = ".json"
)

var GlobalParameters = Params{}

type Params struct {
	NotionTask map[int64]*NotionTaskParams `json:"notion_task,omitempty"`
}

type NotionTaskParams struct {
	NotionTitle       string `json:"notion_title,omitempty"`
	NotionDescription string `json:"notion_description,omitempty"`
	NotionStatus      string `json:"notion_status,omitempty"`
	NotionLang        string `json:"notion_lang,omitempty"`
	NotionBot         string `json:"notion_bot,omitempty"`
}

func UploadParams() {
	var settings *Params
	data, err := os.ReadFile(paramsPath + jsonFormatName)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(data, &settings)
	if err != nil {
		fmt.Println(err)
	}

	if GlobalParameters.NotionTask == nil {
		GlobalParameters.NotionTask = make(map[int64]*NotionTaskParams)
	}

	SaveParams()
}

func SaveParams() {
	data, err := json.MarshalIndent(GlobalParameters, "", "  ")
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile(paramsPath+jsonFormatName, data, 0600); err != nil {
		panic(err)
	}
}

func (p *Params) GetTitle(userID int64) string {
	return p.NotionTask[userID].NotionTitle
}

func (p *Params) UpdateTitle(userID int64, title string) {
	if _, ok := p.NotionTask[userID]; !ok {
		p.NotionTask[userID] = &NotionTaskParams{}
	}

	p.NotionTask[userID].NotionTitle = title
}

func (p *Params) GetDescription(userID int64) string {
	return p.NotionTask[userID].NotionDescription
}

func (p *Params) UpdateDescription(description string, userID int64) {
	if _, ok := p.NotionTask[userID]; !ok {
		p.NotionTask[userID] = &NotionTaskParams{}
	}

	p.NotionTask[userID].NotionDescription = description
}

func (p *Params) GetStatus(userID int64) string {
	return p.NotionTask[userID].NotionStatus
}

func (p *Params) UpdateStatus(userID int64, status string) {
	if _, ok := p.NotionTask[userID]; !ok {
		p.NotionTask[userID] = &NotionTaskParams{}
	}

	p.NotionTask[userID].NotionStatus = status
}

func (p *Params) GetLang(userID int64) string {
	return p.NotionTask[userID].NotionLang
}

func (p *Params) UpdateLang(userID int64, notionLang string) {
	if _, ok := p.NotionTask[userID]; !ok {
		p.NotionTask[userID] = &NotionTaskParams{}
	}

	p.NotionTask[userID].NotionLang = notionLang
}

func (p *Params) GetBot(userID int64) string {
	return p.NotionTask[userID].NotionBot
}

func (p *Params) UpdateBot(userID int64, bot string) {
	if _, ok := p.NotionTask[userID]; !ok {
		p.NotionTask[userID] = &NotionTaskParams{}
	}

	p.NotionTask[userID].NotionBot = bot
}
