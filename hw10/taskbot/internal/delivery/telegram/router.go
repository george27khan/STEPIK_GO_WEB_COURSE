package telegram

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
	"taskbot/internal/domain"
)

type taskService interface {
	GetAll(ctx context.Context) ([]*domain.Task, error)
	Create(ctx context.Context, task domain.Task) (uint64, error)
	AssignTask(ctx context.Context, executor string, executorChatID int64, taskID uint64) (domain.Task, error)
	UnassignTask(ctx context.Context, taskID uint64) error
	GetExecutorChatID(ctx context.Context, taskID uint64) (int64, error)
	GetTask(ctx context.Context, taskID uint64) (*domain.Task, error)
	ResolveTask(ctx context.Context, taskID uint64) error
	GetUserTask(ctx context.Context, username string, role string) ([]*domain.Task, error)
}

type Router struct {
	taskService taskService
	bot         *tgbotapi.BotAPI
}

func NewRouter(task taskService, bot *tgbotapi.BotAPI) *Router {
	return &Router{task, bot}
}

func (r *Router) Route(ctx context.Context) {
	updates := r.bot.ListenForWebhook("/")
	for update := range updates {
		command := update.Message.Text
		if command == "/tasks" {
			r.handlerGetAllTask(ctx, update)
			continue
		} else if strings.HasPrefix(command, "/new") {
			r.handlerCreateTask(ctx, update)
			continue
		} else if strings.HasPrefix(command, "/assign") {
			r.handlerAssignTask(ctx, update)
			continue
		} else if strings.HasPrefix(command, "/unassign") {
			r.handlerUnassignTask(ctx, update)
			continue
		} else if strings.HasPrefix(command, "/resolve") {
			r.handlerResolveTask(ctx, update)
			continue
		} else if strings.HasPrefix(command, "/my") {
			r.handlerGetUserTask(ctx, update)
			continue
		} else if strings.HasPrefix(command, "/owner") {
			r.handlerGetOwnerTask(ctx, update)
			continue
		} else if strings.HasPrefix(command, "/owner") {
			r.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))
			continue
		}
	}
}

func (r *Router) handlerGetAllTask(ctx context.Context, update tgbotapi.Update) {
	tasks, err := r.taskService.GetAll(ctx)
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка при получении задач: %s", err.Error())))
		return
	}
	text := strings.Builder{}
	for _, task := range tasks {
		text.WriteString(fmt.Sprintf("%v. %s by @%s", task.ID, task.Text, task.Author))
		if task.Executor == "" {
			text.WriteString(fmt.Sprintf("\n/assign_%v", task.ID))
		} else if task.Executor == task.Author {
			text.WriteString("\nassignee: я")
			text.WriteString(fmt.Sprintf("\n/unassign_%v /resolve_%v", task.ID, task.ID))
		} else {
			text.WriteString(fmt.Sprintf("\nassignee: ", task.Executor))
		}
		text.WriteString("\n")
	}
	if text.Len() == 0 {
		text.WriteString("Нет задач")
	}
	r.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, text.String()))
}

func (r *Router) handlerCreateTask(ctx context.Context, update tgbotapi.Update) {
	text, _ := strings.CutPrefix(update.Message.Text, "/new ")
	task := domain.Task{
		Text:         text,
		Author:       update.Message.Chat.UserName,
		AuthorChatID: update.Message.Chat.ID,
	}
	id, err := r.taskService.Create(ctx, task)
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка при создании задачи: %s", err.Error())))
	}
	message := fmt.Sprintf("Задача \"%s\" создана, id=%v", task.Text, id)
	r.bot.Send(tgbotapi.NewMessage(task.AuthorChatID, message))
}

func (r *Router) handlerAssignTask(ctx context.Context, update tgbotapi.Update) {
	message := update.Message
	curChatID := message.Chat.ID
	strTaskID, found := strings.CutPrefix(message.Text, "/assign_")
	if !found {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Ошибка определения задачи"))
		return
	}
	taskID, err := strconv.ParseUint(strTaskID, 10, 64)
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Неверный формат задачи"))
		return
	}
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Ошибка при поиске текущего исполнителя"))
		return
	}
	task, err := r.taskService.AssignTask(ctx, message.Chat.UserName, message.Chat.ID, taskID)
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка при назначении задачи: %s", err.Error())))
	}

	messageExecutor := fmt.Sprintf("Задача \"%s\" назначена на вас", task.Text)
	if curChatID == task.AuthorChatID {
		r.bot.Send(tgbotapi.NewMessage(curChatID, messageExecutor))
	} else {
		messageAuthor := fmt.Sprintf("Задача \"%s\" назначена на @%s", task.Text, task.Executor)
		r.bot.Send(tgbotapi.NewMessage(curChatID, messageExecutor))
		r.bot.Send(tgbotapi.NewMessage(task.AuthorChatID, messageAuthor))
	}
}

func (r *Router) handlerUnassignTask(ctx context.Context, update tgbotapi.Update) {
	message := update.Message
	curChatID := message.Chat.ID
	strTaskID, found := strings.CutPrefix(message.Text, "/unassign_")
	if !found {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Ошибка определения задачи"))
		return
	}

	taskID, err := strconv.ParseUint(strTaskID, 10, 64)
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Неверный формат задачи"))
		return
	}

	task, err := r.taskService.GetTask(ctx, taskID)
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Ошибка при поиске текущего исполнителя"))
		return
	}
	if task.Executor != message.Chat.UserName {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Задача не на вас"))
		return
	}

	if err = r.taskService.UnassignTask(ctx, taskID); err != nil {
		r.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка при снятии задачи: %s", err.Error())))
	}
	if curChatID == task.AuthorChatID {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Принято"))
	} else {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Принято"))
		r.bot.Send(tgbotapi.NewMessage(task.AuthorChatID, fmt.Sprintf("Задача \"%s\" осталась без исполнителя", task.Text)))
	}
}

func (r *Router) handlerResolveTask(ctx context.Context, update tgbotapi.Update) {
	message := update.Message
	curChatID := message.Chat.ID
	strTaskID, found := strings.CutPrefix(message.Text, "/resolve_")
	if !found {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Ошибка определения задачи"))
		return
	}

	taskID, err := strconv.ParseUint(strTaskID, 10, 64)
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Неверный формат задачи"))
		return
	}

	task, err := r.taskService.GetTask(ctx, taskID)
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Ошибка при поиске текущего исполнителя"))
		return
	}
	if task.Executor != message.Chat.UserName {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Задача не на вас"))
		return
	}

	if err = r.taskService.ResolveTask(ctx, taskID); err != nil {
		r.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Ошибка при решении задачи: %s", err.Error())))
	}
	if curChatID == task.AuthorChatID {
		r.bot.Send(tgbotapi.NewMessage(curChatID, fmt.Sprintf("Задача \"%s\" выполнена", task.Text)))
	} else {
		r.bot.Send(tgbotapi.NewMessage(curChatID, fmt.Sprintf("Задача \"%s\" выполнена", task.Text)))
		r.bot.Send(tgbotapi.NewMessage(task.AuthorChatID, fmt.Sprintf("Задача \"%s\" выполнена @%s", task.Text, task.Executor)))
	}
}

func (r *Router) handlerGetUserTask(ctx context.Context, update tgbotapi.Update) {
	message := update.Message
	curChatID := message.Chat.ID

	tasks, err := r.taskService.GetUserTask(ctx, message.Chat.UserName, "executor")
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Ошибка при поиске задач на исполнителе"))
		return
	}
	text := strings.Builder{}
	for _, task := range tasks {
		text.WriteString(fmt.Sprintf("%v. %s by @%s", task.ID, task.Text, task.Author))
		text.WriteString(fmt.Sprintf("\n/unassign_%v /resolve_%v", task.ID, task.ID))
		text.WriteString("\n")
	}
	if text.Len() == 0 {
		text.WriteString("Нет задач")
	}
	r.bot.Send(tgbotapi.NewMessage(curChatID, text.String()))
}

func (r *Router) handlerGetOwnerTask(ctx context.Context, update tgbotapi.Update) {
	message := update.Message
	curChatID := message.Chat.ID

	tasks, err := r.taskService.GetUserTask(ctx, message.Chat.UserName, "author")
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Ошибка при поиске задач на исполнителе"))
		return
	}

	text := strings.Builder{}
	for _, task := range tasks {
		text.WriteString(fmt.Sprintf("%v. %s by @%s", task.ID, task.Text, task.Author))
		text.WriteString(fmt.Sprintf("\n/assign_%v", task.ID))
		text.WriteString("\n")
	}
	if text.Len() == 0 {
		text.WriteString("Нет задач")
	}
	r.bot.Send(tgbotapi.NewMessage(curChatID, text.String()))
}
