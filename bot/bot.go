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
	b.Start(ctx)
}
