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
	GetExecutorChatID(ctx context.Context, taskID uint64) (int64, error)
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
		}

		//task := domain.Task{
		//	ChatID: update.Message.Chat.ID,
		//}
		//fmt.Println(update.UpdateID, update.Message.Text, , update.Message.Chat.UserName)

	}
}

func (r *Router) handlerGetAllTask(ctx context.Context, update tgbotapi.Update) {
	tasks, err := r.taskService.GetAll(ctx)
	if err != nil {
		fmt.Println(err)
	}
	message := strings.Builder{}
	for _, task := range tasks {
		message.WriteString(fmt.Sprintf("%v. %s by @%s", task.ID, task.Text, task.Author))
		if task.Executor == "" {
			message.WriteString(fmt.Sprintf("\n/assign_%v", task.ID))
		} else if task.Executor == task.Author {
			message.WriteString("\nassignee: я")
		} else {
			message.WriteString(fmt.Sprintf("\nassignee: ", task.Executor))
		}
		message.WriteString("\n")
	}
	if message.Len() == 0 {
		message.WriteString("Нет задач")
	}
	r.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, message.String()))
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
		fmt.Println(err)
	}
	fmt.Println("id", id)
	message := fmt.Sprintf("Задача \"%s\" создана, id=%v", task.Text, id)
	r.bot.Send(tgbotapi.NewMessage(task.AuthorChatID, message))
}

func (r *Router) handlerAssignTask(ctx context.Context, update tgbotapi.Update) {
	message := update.Message
	curChatID := message.Chat.ID
	strID, found := strings.CutPrefix(message.Text, "/assign_")
	if !found {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Ошибка определения задачи"))
		return
	}
	taskID, err := strconv.ParseUint(strID, 10, 64)
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Неверный формат задачи"))
		return
	}
	prevExecChatID, err := r.taskService.GetExecutorChatID(ctx, taskID)
	if err != nil {
		r.bot.Send(tgbotapi.NewMessage(curChatID, "Ошибка при поиске текущего исполнителя"))
		return
	}
	task, err := r.taskService.AssignTask(ctx, message.Chat.UserName, message.Chat.ID, taskID)
	if err != nil {
		fmt.Println(err)
	}
	messageAuthor := fmt.Sprintf("Задача \"%s\" назначена на %s", task.Text, task.Executor)
	messageExecutor := fmt.Sprintf("Задача \"%s\" назначена на вас", task.Text)
	if prevExecChatID != 0 {
		r.bot.Send(tgbotapi.NewMessage(prevExecChatID, messageAuthor))
	} else {
		r.bot.Send(tgbotapi.NewMessage(task.AuthorChatID, messageAuthor))
	}
	r.bot.Send(tgbotapi.NewMessage(curChatID, messageExecutor))
}
