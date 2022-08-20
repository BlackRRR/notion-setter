package mysql

import (
	"database/sql"
	"github.com/BlackRRR/notion-setter/internal/api/config"
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/services/bot"
	"github.com/bots-empire/base-bot/msgs"
	"github.com/go-sql-driver/mysql"
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
	dataBase.Exec("CREATE TABLE IF NOT EXISTS users (" + config.UserTable + ";")
	dataBase.Exec("CREATE TABLE IF NOT EXISTS tasks (" + config.Tasks + ");")

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

func (r *Repository) DownloadParamsFromDB() {
	rows, err := r.bot.GetDataBase().Query("SELECT * FROM tasks")
	if err != nil {
		log.Printf("error db query download params: %s", err.Error())
	}

	err = readRows(rows)
	if err != nil {
		log.Printf("error read rows download params: %s", err.Error())
	}
}

func readRows(rows *sql.Rows) error {
	var id int64
	var param model.NotionTaskParams
	for rows.Next() {
		err := rows.Scan(
			&id,
			&param.NotionTitle,
			&param.NotionStatus,
			&param.NotionService,
			&param.NotionLang,
			&param.NotionDescription)
		if err != nil {
			return err
		}

		model.GlobalParameters.NotionTask[id] = &param
	}

	return nil
}

func (r *Repository) CreateTaskWithID(id int64) error {
	params := model.NotionTaskParams{
		NotionTitle:       " ",
		NotionDescription: " ",
		NotionStatus:      " ",
		NotionLang:        " ",
		NotionService:     " ",
	}
	_, err := r.bot.GetDataBase().Exec("INSERT INTO tasks VALUES (?, ?, ?, ?, ?, ?)",
		&id,
		&params.NotionTitle,
		&params.NotionStatus,
		&params.NotionService,
		&params.NotionLang,
		&params.NotionDescription,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return nil
		}

		return err
	}

	return nil
}

func (r *Repository) UploadParamsToDB(id int64, params *model.NotionTaskParams) error {
	_, err := r.bot.GetDataBase().Exec(`
UPDATE tasks SET 
                 title = ?, 
                 status = ?, 
                 service = ?, 
                 lang = ?, 
                 description = ? 
             WHERE id = ?`,
		&params.NotionTitle,
		&params.NotionStatus,
		&params.NotionService,
		&params.NotionLang,
		&params.NotionDescription,
		&id)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UploadTitleTODB(id int64, title string) error {
	_, err := r.bot.GetDataBase().Exec("UPDATE tasks SET title = ? WHERE id = ?",
		&title,
		&id)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UploadStatusTODB(id int64, status string) error {
	_, err := r.bot.GetDataBase().Exec("UPDATE tasks SET status = ? WHERE id = ?",
		&status,
		&id)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UploadServiceTODB(id int64, service string) error {
	_, err := r.bot.GetDataBase().Exec("UPDATE tasks SET service = ? WHERE id = ?",
		&service,
		&id)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UploadLangTODB(id int64, lang string) error {
	_, err := r.bot.GetDataBase().Exec("UPDATE tasks SET lang = ? WHERE id = ?",
		&lang,
		&id)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UploadDescriptionTODB(id int64, description string) error {
	_, err := r.bot.GetDataBase().Exec("UPDATE tasks SET description = ? WHERE id = ?",
		&description,
		&id)
	if err != nil {
		return err
	}

	return nil
}
