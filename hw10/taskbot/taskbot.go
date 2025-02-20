package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	router "taskbot/internal/delivery/telegram"
	storage "taskbot/internal/repository/local_storage"
	"taskbot/internal/service"
)

func loadEnv() {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}
}

func startTaskBot(ctx context.Context, httpListenAddr string) error {
	// сюда писать код
	/*
		в этом месте вы стартуете бота,
		стартуете хттп сервер который будет обслуживать этого бота
		инициализируете ваше приложение
		и потом будете обрабатывать входящие сообщения
	*/
	loadEnv()
	// Читаем переменные окружения
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	webhookLink := os.Getenv("WEBHOOK_LINK")
	db := storage.NewTaskStorage()
	taskService := service.NewTaskService(db)
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return err
	}
	bot.SetWebhook(tgbotapi.NewWebhook(webhookLink))
	r := router.NewRouter(taskService, bot)

	go r.Route(ctx)
	fmt.Println("start listen :8081")
	if err = http.ListenAndServe(":8081", nil); err != nil {
		fmt.Println("start server error", err.Error())
		return err
	}
	return nil
}

func main() {
	err := startTaskBot(context.Background(), ":8081")
	if err != nil {
		log.Fatalln(err)
	}
}

// это заглушка чтобы импорт сохранился
func __dummy() {
	tgbotapi.APIEndpoint = "_dummy"
}
