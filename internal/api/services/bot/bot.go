package bot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jomei/notionapi"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	tokensPath     = "./internal/api/config/tokens"
	jsonFormatName = ".json"

	commandsPath            = "assets/commands"
	beginningOfUserLangPath = "assets/language/"
)

var Bot *GlobalBot

type GlobalBot struct {
	BotLang string `json:"bot_lang,omitempty"`

	Bot      *tgbotapi.BotAPI
	Chanel   tgbotapi.UpdatesChannel
	Rdb      *redis.Client
	DataBase *sql.DB
	Notion   *notionapi.Client

	MessageHandler  model.GlobalHandlers
	CallbackHandler model.GlobalHandlers

	Commands map[string]string
	Language map[string]map[string]string

	BotToken    string `json:"bot_token,omitempty"`
	BotLink     string `json:"bot_link,omitempty"`
	NotionToken string `json:"notion_token,omitempty"`
}

func FillBotsConfig() {
	bytes, err := os.ReadFile(tokensPath + jsonFormatName)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &Bot)
	if err != nil {
		log.Fatal(err)
	}
}

func UploadUpdateStatistic() {
	info := &model.UpdateInfo{}
	info.Mu = new(sync.Mutex)
	strStatistic, err := Bot.Rdb.Get("update_statistic").Result()
	if err != nil {
		model.UpdateStatistic = info
		return
	}

	info.Counter, _ = strconv.Atoi(strStatistic)
	model.UpdateStatistic = info
}

func SaveUpdateStatistic() {
	_, err := Bot.Rdb.Set("update_statistic", strconv.Itoa(model.UpdateStatistic.Counter), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func (b *GlobalBot) GetBotLang() string {
	return b.BotLang
}

func (b *GlobalBot) GetBot() *tgbotapi.BotAPI {
	return b.Bot
}

func (b *GlobalBot) GetDataBase() *sql.DB {
	return b.DataBase
}

func (b *GlobalBot) AvailableLang() []string {
	return nil
}

func (b *GlobalBot) GetCurrency() string {
	return ""
}

func (b *GlobalBot) LangText(lang, key string, values ...interface{}) string {
	formatText := b.Language[lang][key]
	return fmt.Sprintf(formatText, values...)
}

func (b *GlobalBot) GetTexts(lang string) map[string]string {
	return b.Language[lang]
}

func (b *GlobalBot) CheckAdmin(userID int64) bool {
	return false
}

func (b *GlobalBot) AdminLang(userID int64) string {
	return ""
}

func (b *GlobalBot) AdminText(adminLang, key string) string {
	return ""
}

func (b *GlobalBot) UpdateBlockedUsers(channel int) {
}

func (b *GlobalBot) GetAdvertURL(userLang string, channel int) string {
	return ""
}

func (b *GlobalBot) GetAdvertText(userLang string, channel int) string {
	return ""
}

func (b *GlobalBot) GetAdvertisingPhoto(lang string, channel int) string {
	return ""
}

func (b *GlobalBot) GetAdvertisingVideo(lang string, channel int) string {
	return ""
}

func (b *GlobalBot) ButtonUnderAdvert() bool {
	return false
}

func (b *GlobalBot) AdvertisingChoice(channel int) string {
	return ""
}

func (b *GlobalBot) BlockUser(userID int64) error {
	return nil
}

func (b *GlobalBot) GetMetrics(metricKey string) *prometheus.CounterVec {
	return nil
}

func (b *GlobalBot) ParseCommandsList() {
	bytes, _ := os.ReadFile(commandsPath + jsonFormatName)
	_ = json.Unmarshal(bytes, &b.Commands)
}

func (b *GlobalBot) ParseLangMap() {
	bytes, _ := os.ReadFile(beginningOfUserLangPath + b.BotLang + jsonFormatName)
	dictionary := make(map[string]string)

	_ = json.Unmarshal(bytes, &dictionary)
	b.Language = make(map[string]map[string]string)
	b.Language[b.BotLang] = dictionary
}

func (b *GlobalBot) GetCommandFromText(message *tgbotapi.Message, userLang string, userID int64) (string, error) {
	searchText := getSearchText(message)
	for key, text := range b.Language[userLang] {
		if text == searchText {
			return b.Commands[key], nil
		}
	}

	command := b.Commands[searchText]
	if command != "" {
		return command, nil
	}

	return "", model.ErrCommandNotConverted
}

func getSearchText(message *tgbotapi.Message) string {
	if message.Command() != "" {
		return strings.Split(message.Text, " ")[0]
	}
	return message.Text
}
