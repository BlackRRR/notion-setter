package message

import (
	"context"
	"github.com/BlackRRR/notion-setter/internal/api/config"
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/redis"
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jomei/notionapi"
)

func (h *MessagesHandlers) StartCommand(s *model.Situation) error {
	text := h.BaseBot.Bot.LangText(s.BotLang, "task_start")
	redis.RdbSetUser(s.User.ID, "main")

	database, err := h.BaseBot.Bot.Notion.Database.Get(context.Background(), config.DatabaseID)
	if err != nil {
		return err
	}

	properties := database.Properties["Bot"].(*notionapi.MultiSelectPropertyConfig)

	var buttons []tgbotapi.InlineKeyboardButton
	var optionLen int

	for i := range properties.MultiSelect.Options {
		data := "/service?" + properties.MultiSelect.Options[i].Name
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

func (h *MessagesHandlers) TaskTitle(s *model.Situation) error {
	title := s.Message.Text
	model.GlobalParameters.UpdateTitle(s.User.ID, title)

	text := h.BaseBot.Bot.LangText(s.BotLang, "task_description")

	err := h.BaseBot.Msgs.NewParseMessage(s.User.ID, text)
	if err != nil {
		return err
	}

	redis.RdbSetUser(s.User.ID, "/task_description")

	return nil
}

func (h *MessagesHandlers) TaskDescription(s *model.Situation) error {
	description := s.Message.Text
	model.GlobalParameters.UpdateDescription(description, s.User.ID)

	text := h.BaseBot.Bot.LangText(s.BotLang, "task_description_added")
	redis.RdbSetUser(s.User.ID, "main")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("task_upload", "/task_upload")),
	).Build(h.BaseBot.Bot.Language[s.BotLang])

	err := h.BaseBot.Msgs.NewParseMarkUpMessage(s.User.ID, &markUp, text)
	if err != nil {
		return err
	}

	return nil
}
