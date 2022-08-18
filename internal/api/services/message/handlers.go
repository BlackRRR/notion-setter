package message

import (
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/redis"
	"github.com/bots-empire/base-bot/msgs"
)

func (h *MessagesHandlers) StartCommand(s *model.Situation) error {
	text := h.BaseBot.Bot.LangText(s.BotLang, "task_start")
	redis.RdbSetUser(s.BotLang, s.User.ID, "main")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("task_bot_referral", "/bot?referral")),
		msgs.NewIlRow(msgs.NewIlDataButton("task_bot_miner", "/bot?miner")),
		msgs.NewIlRow(msgs.NewIlDataButton("task_bot_voice", "/bot?voice")),
		msgs.NewIlRow(msgs.NewIlDataButton("task_bot_youtube", "/bot?youtube")),
		msgs.NewIlRow(msgs.NewIlDataButton("task_bot_tiktok", "/bot?tiktok")),
	).Build(h.BaseBot.Bot.Language[s.BotLang])

	_, err := h.BaseBot.Msgs.NewIDParseMarkUpMessage(s.User.ID, markUp, text)
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

	redis.RdbSetUser(s.BotLang, s.User.ID, "/task_description")

	return nil
}

func (h *MessagesHandlers) TaskDescription(s *model.Situation) error {
	description := s.Message.Text
	model.GlobalParameters.UpdateDescription(description, s.User.ID)

	text := h.BaseBot.Bot.LangText(s.BotLang, "task_description_added")
	redis.RdbSetUser(s.BotLang, s.User.ID, "main")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("task_upload", "/task_upload")),
	).Build(h.BaseBot.Bot.Language[s.BotLang])

	err := h.BaseBot.Msgs.NewParseMarkUpMessage(s.User.ID, &markUp, text)
	if err != nil {
		return err
	}

	return nil
}
