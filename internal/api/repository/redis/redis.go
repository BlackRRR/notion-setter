package redis

import (
	"github.com/BlackRRR/notion-setter/internal/api/services/bot"
	"github.com/go-redis/redis"
	"log"
	"strconv"
)

var (
	redisDefaultAddr = "127.0.0.1:6379"
	emptyLevelName   = "empty"
)

func StartRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisDefaultAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return rdb
}

func RdbSetUser(botLang string, ID int64, level string) {
	userID := userIDToRdb(ID)
	_, err := bot.Bot.Rdb.Set(userID, level, 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func userIDToRdb(userID int64) string {
	return "user:" + strconv.FormatInt(userID, 10)
}

func GetLevel(botLang string, id int64) string {
	userID := userIDToRdb(id)
	have, err := bot.Bot.Rdb.Exists(userID).Result()
	if err != nil {
		log.Println(err)
	}
	if have == 0 {
		return emptyLevelName
	}

	value, err := bot.Bot.Rdb.Get(userID).Result()
	if err != nil {
		log.Println(err)
	}
	return value
}
