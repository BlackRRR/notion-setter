package updates

import (
	"encoding/json"
	"fmt"
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/mysql"
	"github.com/BlackRRR/notion-setter/internal/api/repository/redis"
	"github.com/BlackRRR/notion-setter/internal/api/services/bot"
	"github.com/BlackRRR/notion-setter/internal/api/utils"
	"github.com/BlackRRR/notion-setter/internal/log"
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"runtime/debug"
	"strings"
)

var (
	panicLogger = log.NewDefaultLogger().Prefix("panic cather")

	updatePrintHeader = "updates number: %d    // referral-bot-updates:  %s %s"
	extraneousUpdate  = "extraneous updates"
)

type BaseBot struct {
	Bot  *bot.GlobalBot
	Rep  *mysql.Repository
	Msgs *msgs.Service
}

func NewBaseBotService(bot *bot.GlobalBot, rep *mysql.Repository, msgs *msgs.Service) *BaseBot {
	return &BaseBot{
		Bot:  bot,
		Rep:  rep,
		Msgs: msgs,
	}
}

func (b *BaseBot) checkCallbackQuery(s *model.Situation, logger log.Logger) {
	Handler := bot.Bot.CallbackHandler.
		GetHandler(s.Command)

	if Handler != nil {
		if err := Handler(s); err != nil {
			logger.Warn("error with serve user callback command: %s", err.Error())
			b.smthWentWrong(s.CallbackQuery.Message.Chat.ID, b.Bot.BotLang)
		}
		return
	}

	logger.Warn("get callback data='%s', but they didn't react in any way", s.CallbackQuery.Data)
}

func (b *BaseBot) ActionsWithUpdates(logger log.Logger, sortCentre *utils.Spreader) {
	for update := range b.Bot.Chanel {
		localUpdate := update

		go b.checkUpdate(&localUpdate, logger, sortCentre)
	}
}

func (b *BaseBot) checkUpdate(update *tgbotapi.Update, logger log.Logger, sortCentre *utils.Spreader) {
	defer b.panicCather(update)

	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if update.Message != nil && update.Message.PinnedMessage != nil {
		return
	}

	b.printNewUpdate(update, logger)
	if update.Message != nil && update.Message.From != nil {
		user, err := b.Rep.CheckingTheUser(update.Message)
		if err != nil {
			b.smthWentWrong(update.Message.Chat.ID, b.Bot.BotLang)
			logger.Warn("err with check user: %s", err.Error())
			return
		}

		err = b.Rep.CreateTaskWithID(user.ID)
		if err != nil {
			b.smthWentWrong(update.Message.Chat.ID, b.Bot.BotLang)
			logger.Warn("err with create task with user id: %s, %s", err.Error(), user.ID)
		}

		situation := createSituationFromMsg(b.Bot.BotLang, update.Message, user)

		b.checkMessage(situation, logger, sortCentre)
		return
	}

	if update.CallbackQuery != nil {
		situation, err := b.createSituationFromCallback(b.Bot.BotLang, update.CallbackQuery)
		if err != nil {
			b.smthWentWrong(update.CallbackQuery.Message.Chat.ID, b.Bot.BotLang)
			logger.Warn("err with create situation from callback: %s", err.Error())
			return
		}

		b.checkCallbackQuery(situation, logger)
		return
	}
}

func (b *BaseBot) printNewUpdate(update *tgbotapi.Update, logger log.Logger) {
	model.UpdateStatistic.Mu.Lock()
	defer model.UpdateStatistic.Mu.Unlock()

	model.UpdateStatistic.Counter++
	bot.SaveUpdateStatistic()

	model.HandleUpdates.WithLabelValues(
		b.Bot.BotLink,
		b.Bot.BotLang,
	).Inc()

	if update.Message != nil {
		if update.Message.Text != "" {
			logger.Info(updatePrintHeader,
				model.UpdateStatistic.Counter,
				b.Bot.BotLang,
				update.Message.Text,
			)
			return
		}
	}

	if update.CallbackQuery != nil {
		logger.Info(updatePrintHeader,
			model.UpdateStatistic.Counter,
			b.Bot.BotLang,
			update.CallbackQuery.Data,
		)
		return
	}

	logger.Info(updatePrintHeader,
		model.UpdateStatistic.Counter,
		b.Bot.BotLang,
		extraneousUpdate,
	)
}

func createSituationFromMsg(botLang string, message *tgbotapi.Message, user *model.User) *model.Situation {
	return &model.Situation{
		Message: message,
		BotLang: botLang,
		User:    user,
		Params: &model.Parameters{
			Level: redis.GetLevel(user.ID),
		},
	}
}

func (b *BaseBot) createSituationFromCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery) (*model.Situation, error) {
	user, err := b.Rep.GetUser(callbackQuery.From.ID)
	if err != nil {
		return nil, err
	}

	return &model.Situation{
		CallbackQuery: callbackQuery,
		BotLang:       botLang,
		User:          user,
		Command:       strings.Split(callbackQuery.Data, "?")[0],
		Params: &model.Parameters{
			Level: redis.GetLevel(callbackQuery.From.ID),
		},
	}, nil
}

func (b *BaseBot) checkMessage(situation *model.Situation, logger log.Logger, sortCentre *utils.Spreader) {
	if situation.Command == "" {
		situation.Command, situation.Err = b.Bot.GetCommandFromText(
			situation.Message, b.Bot.BotLang, situation.User.ID)
	}

	if situation.Err == nil {
		handler := bot.Bot.MessageHandler.
			GetHandler(situation.Command)

		if handler != nil {
			sortCentre.ServeHandler(handler, situation, func(err error) {
				text := fmt.Sprintf("%s // %s // error with serve user msg command: %s",
					b.Bot.BotLang,
					b.Bot.BotLink,
					err.Error(),
				)
				b.Msgs.SendNotificationToDeveloper(text, false)

				logger.Warn(text)
				b.smthWentWrong(situation.Message.Chat.ID, b.Bot.BotLang)
			})
			return
		}
	}

	situation.Command = strings.Split(situation.Params.Level, "?")[0]

	handler := bot.Bot.MessageHandler.
		GetHandler(situation.Command)

	if handler != nil {
		sortCentre.ServeHandler(handler, situation, func(err error) {
			text := fmt.Sprintf("%s // %s // error with serve user level command: %s",
				b.Bot.BotLang,
				b.Bot.BotLink,
				err.Error(),
			)
			b.Msgs.SendNotificationToDeveloper(text, false)

			logger.Warn(text)
			b.smthWentWrong(situation.Message.Chat.ID, b.Bot.BotLang)
		})
		return
	}

	b.smthWentWrong(situation.Message.Chat.ID, b.Bot.BotLang)
	if situation.Err != nil {
		logger.Info(situation.Err.Error())
	}
}

func (b *BaseBot) smthWentWrong(chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, b.Bot.LangText(lang, "user_level_not_defined"))
	_ = b.Msgs.SendMsgToUser(msg)
}

func (b *BaseBot) panicCather(update *tgbotapi.Update) {
	msg := recover()
	if msg == nil {
		return
	}

	panicText := fmt.Sprintf("%s // %s\npanic in backend: message = %s\n%s",
		b.Bot.BotLang,
		b.Bot.BotLink,
		msg,
		string(debug.Stack()),
	)
	panicLogger.Warn(panicText)

	b.Msgs.SendNotificationToDeveloper(panicText, false)

	data, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		return
	}

	b.Msgs.SendNotificationToDeveloper(string(data), false)
}
