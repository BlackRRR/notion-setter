package updates

import (
	"github.com/BlackRRR/notion-setter/internal/api/model"
	"github.com/BlackRRR/notion-setter/internal/api/repository/mysql"
)

type CallBackHandlers struct {
	Handlers map[string]model.Handler
	BaseBot  *BaseBot
	MySqlRep *mysql.Repository
}

func (h *CallBackHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *CallBackHandlers) Init() {
	//Money command
	h.OnCommand("/service", h.BaseBot.TaskService)
	h.OnCommand("/lang", h.BaseBot.TaskLang)
	h.OnCommand("/status", h.BaseBot.TaskStatus)
	h.OnCommand("/task_upload", h.BaseBot.TaskUpload)
	h.OnCommand("/back_to_start", h.BaseBot.Back)
	h.OnCommand("/back_to_desc", h.BaseBot.BackToDesc)
	h.OnCommand("/skip_description", h.BaseBot.SkipDescription)
}

func (h *CallBackHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}
