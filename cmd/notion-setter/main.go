package main

import (
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/mysql"
	"github.com/BlackRRR/notion-setter/internal/api/repository/redis"
	"github.com/BlackRRR/notion-setter/internal/api/services"
	"github.com/BlackRRR/notion-setter/internal/api/services/bot"
	"github.com/BlackRRR/notion-setter/internal/api/services/callback"
	"github.com/BlackRRR/notion-setter/internal/api/services/message"
	"github.com/BlackRRR/notion-setter/internal/api/utils"
	"github.com/BlackRRR/notion-setter/internal/log"
	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jomei/notionapi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())

	logger := log.NewDefaultLogger().Prefix("notion-creator Bot")
	log.PrintLogo("notion-creator Bot", []string{"FFD700"})

	model.UploadParams()
	bot.FillBotsConfig()

	//go startPrometheusHandler(logger)

	srvs := startAllBot(logger)
	bot.UploadUpdateStatistic()

	startHandlers(srvs, logger)
}

func startAllBot(log log.Logger) *services.BaseBot {
	globalBot := bot.Bot
	startBot(globalBot, log)
	startNotionClient(globalBot)
	//1418862576, -1001736803459
	service := msgs.NewService(globalBot, []int64{872383555})

	rep := mysql.NewRepository(globalBot, service)
	baseBot := services.NewBaseBotService(globalBot, rep, service)

	globalBot.MessageHandler = NewMessagesHandler(baseBot, rep)
	globalBot.CallbackHandler = NewCallbackHandler(baseBot, rep)

	log.Ok("All bots is running")
	return baseBot
}

func startBot(b *bot.GlobalBot, log log.Logger) {
	var err error
	b.Bot, err = tgbotapi.NewBotAPI(b.BotToken)
	if err != nil {
		log.Fatal("error start bot: %s", err.Error())
	}

	u := tgbotapi.NewUpdate(0)

	b.Chanel = b.Bot.GetUpdatesChan(u)

	b.Rdb = redis.StartRedis()
	b.DataBase = mysql.UploadDataBase(b.BotLang)

	b.ParseCommandsList()
	b.ParseLangMap()
}

func startNotionClient(b *bot.GlobalBot) {
	b.Notion = notionapi.NewClient(notionapi.Token(b.NotionToken))
}

func startPrometheusHandler(logger log.Logger) {
	http.Handle("/metrics", promhttp.Handler())
	logger.Ok("Metrics can be read from %s port", "7011")
	metricErr := http.ListenAndServe(":7012", nil)
	if metricErr != nil {
		logger.Fatal("metrics stoped by metricErr: %s\n", metricErr.Error())
	}
}

func startHandlers(baseBot *services.BaseBot, logger log.Logger) {
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func(handler *services.BaseBot, wg *sync.WaitGroup) {
		defer wg.Done()
		handler.ActionsWithUpdates(logger, utils.NewSpreader(time.Minute))
	}(baseBot, wg)

	baseBot.Msgs.SendNotificationToDeveloper("Bot are restart", false)

	logger.Ok("All message are running")

	wg.Wait()
}

func NewMessagesHandler(baseBot *services.BaseBot, repository *mysql.Repository) *message.MessagesHandlers {
	handle := message.MessagesHandlers{
		Handlers: map[string]model.Handler{},
		BaseBot:  baseBot,
		MySqlRep: repository,
	}

	handle.Init()
	return &handle
}

func NewCallbackHandler(baseBot *services.BaseBot, repository *mysql.Repository) *callback.CallBackHandlers {
	handle := callback.CallBackHandlers{
		Handlers: map[string]model.Handler{},
		BaseBot:  baseBot,
		MySqlRep: repository,
	}

	handle.Init()
	return &handle
}
