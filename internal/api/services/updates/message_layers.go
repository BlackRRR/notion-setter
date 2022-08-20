package updates

import (
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/mysql"
)

type MessagesHandlers struct {
	Handlers map[string]model.Handler
	BaseBot  *BaseBot
	MySqlRep *mysql.Repository
}

func (h *MessagesHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *MessagesHandlers) Init() {
	//Start command
	h.OnCommand("/start", h.BaseBot.StartCommand)
	h.OnCommand("/task_title", h.BaseBot.TaskTitle)
	h.OnCommand("/task_description", h.BaseBot.TaskDescription)

}

func (h *MessagesHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}
