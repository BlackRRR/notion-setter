package callback

import (
	"context"
	"github.com/BlackRRR/notion-setter/internal/api/config"
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/redis"
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jomei/notionapi"
	"strings"
)

func (h *CallBackHandlers) TaskService(s *model.Situation) error {
	data := strings.Split(s.CallbackQuery.Data, "?")
	service := data[1]
	model.GlobalParameters.UpdateBot(s.User.ID, service)

	text := h.BaseBot.Bot.LangText(s.BotLang, "task_lang")
	redis.RdbSetUser(s.User.ID, "main")

	database, err := h.BaseBot.Bot.Notion.Database.Get(context.Background(), config.DatabaseID)
	if err != nil {
		return err
	}

	properties := database.Properties["Bot Lang"].(*notionapi.MultiSelectPropertyConfig)

	var buttons []tgbotapi.InlineKeyboardButton
	var optionLen int

	for i := range properties.MultiSelect.Options {
		data := "/lang?" + properties.MultiSelect.Options[i].Name
		button := tgbotapi.InlineKeyboardButton{
			Text:         properties.MultiSelect.Options[i].Name,
			CallbackData: &data,
		}
		buttons = append(
			buttons,
			button)
		optionLen++

	}

	var counter int
	var markUp tgbotapi.InlineKeyboardMarkup

	if markUp.InlineKeyboard == nil {
		markUp.InlineKeyboard = make([][]tgbotapi.InlineKeyboardButton, optionLen/3+1)
	}

	for i := range buttons {
		if i == 0 {
			markUp.InlineKeyboard[counter] = append(markUp.InlineKeyboard[counter], buttons[i])
			continue
		}

		if i%3 == 0 {
			counter++
		}

		markUp.InlineKeyboard[counter] = append(markUp.InlineKeyboard[counter], buttons[i])
	}

	_, err = h.BaseBot.Msgs.NewIDParseMarkUpMessage(s.User.ID, markUp, text)
	if err != nil {
		return err
	}

	return nil
}

func (h *CallBackHandlers) TaskLang(s *model.Situation) error {
	data := strings.Split(s.CallbackQuery.Data, "?")
	lang := data[1]
	model.GlobalParameters.UpdateLang(s.User.ID, lang)

	text := h.BaseBot.Bot.LangText(s.BotLang, "task_status")
	redis.RdbSetUser(s.User.ID, "main")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("task_status_no_status", "/status? ")),
		msgs.NewIlRow(msgs.NewIlDataButton("task_status_critical", "/status?Critical")),
	).Build(h.BaseBot.Bot.Language[s.BotLang])

	_, err := h.BaseBot.Msgs.NewIDParseMarkUpMessage(s.User.ID, markUp, text)
	if err != nil {
		return err
	}

	return nil
}

func (h *CallBackHandlers) TaskStatus(s *model.Situation) error {
	data := strings.Split(s.CallbackQuery.Data, "?")
	status := data[1]
	model.GlobalParameters.UpdateStatus(s.User.ID, status)

	text := h.BaseBot.Bot.LangText(s.BotLang, "task_title")

	err := h.BaseBot.Msgs.NewParseMessage(s.User.ID, text)
	if err != nil {
		return err
	}

	redis.RdbSetUser(s.User.ID, "/task_title")
	return nil
}

func (h *CallBackHandlers) TaskUpload(s *model.Situation) error {
	redis.RdbSetUser(s.User.ID, "main")

	page, err := h.BaseBot.Bot.Notion.Page.Create(context.Background(), &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       "database_id",
			DatabaseID: config.DatabaseID,
		},
		Properties: notionapi.Properties{
			"Status": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: model.GlobalParameters.GetStatus(s.User.ID),
				},
			},
			"Bot": notionapi.MultiSelectProperty{
				MultiSelect: []notionapi.Option{
					{Name: model.GlobalParameters.GetBot(s.User.ID)}},
			},
			"Bot Lang": notionapi.MultiSelectProperty{
				MultiSelect: []notionapi.Option{
					{Name: model.GlobalParameters.GetLang(s.User.ID)}},
			},
			"Name": notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{Text: notionapi.Text{Content: model.GlobalParameters.GetTitle(s.User.ID)}},
				},
			},
		},
		Children: []notionapi.Block{notionapi.Heading2Block{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeHeading2,
			},
			Heading2: notionapi.Heading{
				RichText: []notionapi.RichText{{
					Text: notionapi.Text{
						Content: model.GlobalParameters.GetDescription(s.User.ID),
					},
				}},
			},
		},
		},
	})

	if err != nil {
		return err
	}

	text := h.BaseBot.Bot.LangText(s.BotLang, "task_uploaded", page.URL)
	model.SaveParams()
	return h.BaseBot.Msgs.NewParseMessage(s.User.ID, text)
}
