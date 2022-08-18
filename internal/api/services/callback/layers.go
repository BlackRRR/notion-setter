package callback

import (
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/mysql"
	"github.com/BlackRRR/notion-setter/internal/api/services"
)

type CallBackHandlers struct {
	Handlers map[string]model.Handler
	BaseBot  *services.BaseBot
	MySqlRep *mysql.Repository
}

func (h *CallBackHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *CallBackHandlers) Init() {
	//Money command
	h.OnCommand("/service", h.TaskService)
	h.OnCommand("/lang", h.TaskLang)
	h.OnCommand("/status", h.TaskStatus)
	h.OnCommand("/task_upload", h.TaskUpload)

}

func (h *CallBackHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}
