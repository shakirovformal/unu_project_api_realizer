package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/shakirovformal/unu_project_api_realizer/bot"
)

func __init__() {

}

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Ошибка загрузки файла .env: %v", err)
	}

	bot.Bot()
}
