package bot

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
)

func Bot() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(os.Getenv("TG_TOKEN"), opts...)
	if err != nil {
		panic(err)
	}
	slog.Info("BOT STARTED")
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, welcomeMessage)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, helpMessage)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeExact, checkBalance)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/get_folders_id", bot.MatchTypeExact, getFoldersId)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/create_folder", bot.MatchTypeExact, createFolder)

	b.RegisterHandler(bot.HandlerTypeMessageText, "/delete_folder", bot.MatchTypeExact, deleteFolder)

	b.RegisterHandler(bot.HandlerTypeMessageText, "/create_task", bot.MatchTypeExact, createTask)
	//TODO: Реализовать функцию удаления задачи
	// b.RegisterHandler(bot.HandlerTypeMessageText, "/delete_taskr", bot.MatchTypeExact, deleteTask)

	b.Start(ctx)
}
