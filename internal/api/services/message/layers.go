package message

import (
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/mysql"
	"github.com/BlackRRR/notion-setter/internal/api/services"
)

type MessagesHandlers struct {
	Handlers map[string]model.Handler
	BaseBot  *services.BaseBot
	MySqlRep *mysql.Repository
}

func (h *MessagesHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *MessagesHandlers) Init() {
	//Start command
	h.OnCommand("/start", h.StartCommand)
	h.OnCommand("/task_title", h.TaskTitle)
	h.OnCommand("/task_description", h.TaskDescription)

}

func (h *MessagesHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}
