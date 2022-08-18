package mysql

import (
	"database/sql"
	"github.com/BlackRRR/notion-setter/internal/api/config"
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/services/bot"
	"github.com/bots-empire/base-bot/msgs"
	_ "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
)

const dbDriver = "mysql"

type Repository struct {
	bot  *bot.GlobalBot
	msgs *msgs.Service
}

func NewRepository(bot *bot.GlobalBot, msgs *msgs.Service) *Repository {
	return &Repository{
		bot:  bot,
		msgs: msgs,
	}
}

func UploadDataBase(dbLang string) *sql.DB {
	dataBase, err := sql.Open(dbDriver, config.DBconfig.User+config.DBconfig.Password+"@/")
	if err != nil {
		log.Fatalf("Failed open database: %s\n", err.Error())
	}

	dataBase.Exec("CREATE DATABASE IF NOT EXISTS " + config.DBconfig.Names[dbLang] + ";")
	dataBase.Exec("USE " + config.DBconfig.Names[dbLang] + ";")
	dataBase.Exec("CREATE TABLE IF NOT EXISTS users (" + config.UserTable + ");")

	dataBase.Close()

	dataBase, err = sql.Open(dbDriver, config.DBconfig.User+config.DBconfig.Password+"@/"+config.DBconfig.Names[dbLang])
	if err != nil {
		log.Fatalf("Failed open database: %s\n", err.Error())
	}

	err = dataBase.Ping()
	if err != nil {
		log.Fatalf("Failed upload database: %s\n", err.Error())
	}

	return dataBase
}

func (r *Repository) CheckingTheUser(message *tgbotapi.Message) (*model.User, error) {
	rows, err := r.bot.GetDataBase().Query(`
SELECT * FROM users 
	WHERE id = ?;`,
		message.From.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get user")
	}

	users, err := ReadUsers(rows)
	if err != nil {
		return nil, errors.Wrap(err, "read user")
	}

	switch len(users) {
	case 0:
		user := createSimpleUser(message)
		if err := r.addNewUser(user); err != nil {
			return nil, errors.Wrap(err, "add new user")
		}
		return user, nil
	case 1:
		return users[0], nil
	default:
		return nil, model.ErrFoundTwoUsers
	}
}

func (r *Repository) addNewUser(u *model.User) error {
	_, err := r.bot.GetDataBase().Exec(`INSERT INTO users VALUES (?);`, u.ID)
	if err != nil {
		return errors.Wrap(err, "insert new user")
	}

	_ = r.msgs.SendSimpleMsg(u.ID, r.bot.LangText(r.bot.BotLang, "start_text"))

	return nil
}

func createSimpleUser(message *tgbotapi.Message) *model.User {
	return &model.User{
		ID: message.From.ID,
	}
}

func (r *Repository) GetUser(id int64) (*model.User, error) {
	rows, err := r.bot.GetDataBase().Query(`
SELECT * FROM users
	WHERE id = ?;`,
		id)
	if err != nil {
		return nil, err
	}

	users, err := ReadUsers(rows)
	if err != nil || len(users) == 0 {
		return nil, model.ErrUserNotFound
	}
	return users[0], nil
}

func ReadUsers(rows *sql.Rows) ([]*model.User, error) {
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		user := &model.User{}

		if err := rows.Scan(
			&user.ID,
		); err != nil {
			return nil, errors.Wrap(err, model.ErrScanSqlRow.Error())
		}

		users = append(users, user)
	}

	return users, nil
}
