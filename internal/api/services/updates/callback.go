package updates

import (
	"context"
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/redis"
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jomei/notionapi"
	"log"
	"strings"
)

func (b *BaseBot) TaskService(s *model.Situation) error {
	data := strings.Split(s.CallbackQuery.Data, "?")
	service := data[1]
	if service == " " {
		model.GlobalParameters.UpdateService(s.User.ID, "")
		err := b.Rep.UploadServiceTODB(s.User.ID, "")
		if err != nil {
			return err
		}
	} else {
		model.GlobalParameters.UpdateService(s.User.ID, service)
		err := b.Rep.UploadServiceTODB(s.User.ID, service)
		if err != nil {
			return err
		}
	}

	text := b.Bot.LangText(s.BotLang, "task_lang")
	redis.RdbSetUser(s.User.ID, "main")

	database, err := b.Bot.Notion.Database.Get(context.Background(), notionapi.DatabaseID(b.Bot.NotionDatabase))
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

	skipChoose := "/lang? "
	markUp.InlineKeyboard = append(markUp.InlineKeyboard, []tgbotapi.InlineKeyboardButton{{
		Text:         b.Bot.LangText(s.BotLang, "skip_choose"),
		CallbackData: &skipChoose,
	},
	})

	backData := "/back_to_start"
	markUp.InlineKeyboard = append(markUp.InlineKeyboard, []tgbotapi.InlineKeyboardButton{{
		Text:         b.Bot.LangText(s.BotLang, "back_to_task_start"),
		CallbackData: &backData,
	},
	})

	redis.RdbSetMessageID(s.User.ID, s.CallbackQuery.Message.MessageID)

	return b.Msgs.NewEditMarkUpMessage(s.User.ID, s.CallbackQuery.Message.MessageID, &markUp, text)
}

func (b *BaseBot) Back(s *model.Situation) error {
	cfg := tgbotapi.DeleteMessageConfig{
		ChatID:    s.User.ID,
		MessageID: redis.GetMsgID(s.User.ID),
	}

	_, err := b.Bot.Bot.Request(cfg)
	if err != nil {
		return err
	}

	return b.StartCommand(s)
}

func (b *BaseBot) TaskLang(s *model.Situation) error {
	data := strings.Split(s.CallbackQuery.Data, "?")
	lang := data[1]

	if lang == " " {
		model.GlobalParameters.UpdateLang(s.User.ID, "")
		err := b.Rep.UploadLangTODB(s.User.ID, "")
		if err != nil {
			return err
		}
	} else {
		model.GlobalParameters.UpdateLang(s.User.ID, lang)
		err := b.Rep.UploadLangTODB(s.User.ID, lang)
		if err != nil {
			return err
		}
	}

	text := b.Bot.LangText(s.BotLang, "task_status")
	redis.RdbSetUser(s.User.ID, "main")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("task_status_no_status", "/status? ")),
		msgs.NewIlRow(msgs.NewIlDataButton("task_status_critical", "/status?Critical")),
		msgs.NewIlRow(msgs.NewIlDataButton("back_to_task_lang", "/service?"+model.GlobalParameters.GetService(s.User.ID))),
	).Build(b.Bot.Language[s.BotLang])

	return b.Msgs.NewEditMarkUpMessage(s.User.ID, s.CallbackQuery.Message.MessageID, &markUp, text)
}

func (b *BaseBot) TaskStatus(s *model.Situation) error {
	data := strings.Split(s.CallbackQuery.Data, "?")
	status := data[1]
	model.GlobalParameters.UpdateStatus(s.User.ID, status)
	b.Rep.UploadStatusTODB(s.User.ID, status)

	text := b.Bot.LangText(s.BotLang, "task_title")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlDataButton("back_to_task_status", "/lang?"+model.GlobalParameters.GetLang(s.User.ID))),
	).Build(b.Bot.Language[s.BotLang])

	redis.RdbSetUser(s.User.ID, "/task_title")
	redis.RdbSetMessageID(s.User.ID, s.CallbackQuery.Message.MessageID)

	return b.Msgs.NewEditMarkUpMessage(s.User.ID, s.CallbackQuery.Message.MessageID, &markUp, text)
}

func (b *BaseBot) TaskUpload(s *model.Situation) error {
	redis.RdbSetUser(s.User.ID, "main")

	page, err := b.Bot.Notion.Page.Create(context.Background(), &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       "database_id",
			DatabaseID: notionapi.DatabaseID(b.Bot.NotionDatabase),
		},
		Properties: b.Properties(s),
		Children:   NewBlock(model.GlobalParameters.GetDescription(s.User.ID)),
	})

	if err != nil {
		return err
	}

	cfg := tgbotapi.DeleteMessageConfig{
		ChatID:    s.User.ID,
		MessageID: redis.GetMsgID(s.User.ID),
	}

	_, err = b.Bot.Bot.Request(cfg)
	if err != nil {
		return err
	}

	text := b.Bot.LangText(s.BotLang, "task_uploaded", page.URL)
	return b.Msgs.NewParseMessage(s.User.ID, text)
}

func SelectProperty(key string) *notionapi.SelectProperty {
	return &notionapi.SelectProperty{
		Select: notionapi.Option{
			Name: key,
		},
	}
}

func MultiSelectProperty(key string) *notionapi.MultiSelectProperty {
	return &notionapi.MultiSelectProperty{
		MultiSelect: []notionapi.Option{{
			Name: key,
		}},
	}
}

func TitleProperty(key string) *notionapi.TitleProperty {
	return &notionapi.TitleProperty{
		Title: []notionapi.RichText{{
			Text: notionapi.Text{
				Content: key,
			},
		}},
	}
}

func NewBlock(key string) []notionapi.Block {
	return []notionapi.Block{
		notionapi.Heading2Block{
			BasicBlock: notionapi.BasicBlock{
				Object: notionapi.ObjectTypeBlock,
				Type:   notionapi.BlockTypeHeading2,
			},
			Heading2: notionapi.Heading{
				RichText: []notionapi.RichText{{
					Text: notionapi.Text{
						Content: key,
					},
				}},
			},
		},
	}
}

func (b *BaseBot) Properties(s *model.Situation) notionapi.Properties {
	database, err := b.Bot.Notion.Database.Get(context.Background(), notionapi.DatabaseID(b.Bot.NotionDatabase))
	if err != nil {
		log.Println(err)
	}

	properties := notionapi.Properties{}

	for i := range database.Properties {
		switch i {
		case "Bot Lang":
			if model.GlobalParameters.GetLang(s.User.ID) == "" {

			} else {
				properties[i] = MultiSelectProperty(model.GlobalParameters.GetLang(s.User.ID))
			}
		case "Service":
			properties[i] = MultiSelectProperty(model.GlobalParameters.GetService(s.User.ID))
		case "Name":
			properties[i] = TitleProperty(model.GlobalParameters.GetTitle(s.User.ID))
		case "Status":
			properties[i] = SelectProperty(model.GlobalParameters.GetStatus(s.User.ID))
		}
	}

	return properties
}
