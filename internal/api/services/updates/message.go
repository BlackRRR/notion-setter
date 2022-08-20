package updates

import (
	"context"
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/redis"
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jomei/notionapi"
)

func (b *BaseBot) StartCommand(s *model.Situation) error {
	text := b.Bot.LangText(s.BotLang, "task_start")
	redis.RdbSetUser(s.User.ID, "main")

	database, err := b.Bot.Notion.Database.Get(context.Background(), notionapi.DatabaseID(b.Bot.NotionDatabase))
	if err != nil {
		return err
	}

	properties := database.Properties["Service"].(*notionapi.MultiSelectPropertyConfig)

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

	_, err = b.Msgs.NewIDParseMarkUpMessage(s.User.ID, markUp, text)
	if err != nil {
		return err
	}

	return nil
}

func (b *BaseBot) TaskTitle(s *model.Situation) error {
	title := s.Message.Text
	model.GlobalParameters.UpdateTitle(s.User.ID, title)
	err := b.Rep.UploadTitleTODB(s.User.ID, title)
	if err != nil {
		return err
	}

	text := b.Bot.LangText(s.BotLang, "task_description")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("skip_choose", "/skip_description")),
		msgs.NewIlRow(msgs.NewIlDataButton("back_to_title", "/status?"+model.GlobalParameters.GetStatus(s.User.ID))),
	).Build(b.Bot.Language[s.BotLang])

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      s.User.ID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	send, err := b.Bot.Bot.Send(msg)
	if err != nil {
		return err
	}

	redis.RdbSetUser(s.User.ID, "/task_description")

	cfg := tgbotapi.DeleteMessageConfig{
		ChatID:    s.User.ID,
		MessageID: redis.GetMsgID(s.User.ID),
	}

	_, err = b.Bot.Bot.Request(cfg)
	if err != nil {
		return err
	}

	redis.RdbSetMessageID(s.User.ID, send.MessageID)

	return nil
}

func (b *BaseBot) TaskDescription(s *model.Situation) error {
	description := s.Message.Text
	model.GlobalParameters.UpdateDescription(description, s.User.ID)
	err := b.Rep.UploadDescriptionTODB(s.User.ID, description)
	if err != nil {
		return err
	}

	var status string
	if model.GlobalParameters.GetStatus(s.User.ID) == " " {
		status = "No status"
	} else {
		status = model.GlobalParameters.GetStatus(s.User.ID)
	}

	text := b.Bot.LangText(s.BotLang,
		"task_was_created",
		model.GlobalParameters.GetTitle(s.User.ID),
		status,
		model.GlobalParameters.GetService(s.User.ID),
		model.GlobalParameters.GetLang(s.User.ID),
		model.GlobalParameters.GetDescription(s.User.ID))

	err = b.Rep.UploadParamsToDB(s.User.ID, &model.NotionTaskParams{
		NotionTitle:       model.GlobalParameters.GetTitle(s.User.ID),
		NotionStatus:      status,
		NotionService:     model.GlobalParameters.GetService(s.User.ID),
		NotionLang:        model.GlobalParameters.GetLang(s.User.ID),
		NotionDescription: model.GlobalParameters.GetDescription(s.User.ID),
	})
	if err != nil {
		return err
	}

	redis.RdbSetUser(s.User.ID, "main")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("task_upload", "/task_upload")),
		msgs.NewIlRow(msgs.NewIlDataButton("back_to_description", "/back_to_desc")),
	).Build(b.Bot.Language[s.BotLang])

	cfg := tgbotapi.DeleteMessageConfig{
		ChatID:    s.User.ID,
		MessageID: redis.GetMsgID(s.User.ID),
	}

	_, err = b.Bot.Bot.Request(cfg)
	if err != nil {
		return err
	}

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      s.User.ID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	send, err := b.Bot.Bot.Send(msg)
	if err != nil {
		return err
	}

	redis.RdbSetMessageID(s.User.ID, send.MessageID)

	return nil
}

func (b *BaseBot) SkipDescription(s *model.Situation) error {
	s.Message = &tgbotapi.Message{
		From: &tgbotapi.User{
			ID: s.User.ID,
		},
		Text: model.GlobalParameters.GetTitle(s.User.ID),
	}

	return b.TaskDescription(s)
}

func (b *BaseBot) BackToDesc(s *model.Situation) error {
	s.Message = &tgbotapi.Message{
		From: &tgbotapi.User{
			ID: s.User.ID,
		},
		Text: model.GlobalParameters.GetDescription(s.User.ID),
	}

	return b.TaskTitle(s)
}
