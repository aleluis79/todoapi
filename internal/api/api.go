package api

import (
	"example.com/todoapi/internal/auth"
	"example.com/todoapi/internal/config"
	"example.com/todoapi/pkg/todos"
	"github.com/gin-gonic/gin"
)

type API struct {
	config config.Config
}

func NewTodosAPI(c config.Config) *API {
	return &API{config: c}
}

// @Summary      Get Todo
// @Description  Get Todo
// @Tags         Todos
// @Accept       json
// @Produce      json
// @success 	 200   {object} todos.Todo  "Todo"
// @Failure      404   {string} string      "Not found"
// @Router       /api/todos [get]
// @Security ApiKeyAuth
func (api *API) Get(c *gin.Context) {

	if auth.Verify(c) {
		var todo = todos.Todo{
			ID:          1,
			Title:       "Prueba",
			Description: "Prueba",
			Completed:   false,
		}
		c.JSON(200, todo)
	}
}
