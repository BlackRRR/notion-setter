package model

import "fmt"

var GlobalParameters = &Params{make(map[int64]*NotionTaskParams)}

type Params struct {
	NotionTask map[int64]*NotionTaskParams `json:"notion_task,omitempty"`
}

type NotionTaskParams struct {
	NotionTitle       string `json:"notion_title,omitempty"`
	NotionStatus      string `json:"notion_status,omitempty"`
	NotionService     string `json:"notion_bot,omitempty"`
	NotionLang        string `json:"notion_lang,omitempty"`
	NotionDescription string `json:"notion_description,omitempty"`
}

func (p *Params) GetTitle(userID int64) string {
	return p.NotionTask[userID].NotionTitle
}

func (p *Params) UpdateTitle(userID int64, title string) {
	if _, ok := p.NotionTask[userID]; !ok {
		p.NotionTask[userID] = &NotionTaskParams{
			NotionTitle: title,
		}
	}

	p.NotionTask[userID].NotionTitle = title
}

func (p *Params) GetDescription(userID int64) string {
	return p.NotionTask[userID].NotionDescription
}

func (p *Params) UpdateDescription(description string, userID int64) {
	if _, ok := p.NotionTask[userID]; !ok {
		p.NotionTask[userID] = &NotionTaskParams{
			NotionDescription: description,
		}
	}

	p.NotionTask[userID].NotionDescription = description
}

func (p *Params) GetStatus(userID int64) string {
	return p.NotionTask[userID].NotionStatus
}

func (p *Params) UpdateStatus(userID int64, status string) {
	if _, ok := p.NotionTask[userID]; !ok {
		p.NotionTask[userID] = &NotionTaskParams{
			NotionStatus: status,
		}
	}

	p.NotionTask[userID].NotionStatus = status
}

func (p *Params) GetLang(userID int64) string {
	return p.NotionTask[userID].NotionLang
}

func (p *Params) UpdateLang(userID int64, notionLang string) {
	if _, ok := p.NotionTask[userID]; !ok {
		p.NotionTask[userID] = &NotionTaskParams{
			NotionLang: notionLang,
		}
	}

	p.NotionTask[userID].NotionLang = notionLang
}

func (p *Params) GetService(userID int64) string {
	return p.NotionTask[userID].NotionService
}

func (p *Params) UpdateService(userID int64, service string) {
	if _, ok := p.NotionTask[userID]; !ok {
		p.NotionTask[userID] = &NotionTaskParams{
			NotionService: service,
		}
	}

	p.NotionTask[userID].NotionService = service
	fmt.Println(service)
}
