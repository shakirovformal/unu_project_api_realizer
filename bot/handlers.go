package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/shakirovformal/unu_project_api_realizer/api"
)

type UserState struct {
	State     string
	Data      map[string]interface{}
	CreatedAt time.Time
	Command   string
}

const (
	STATE_WAIT_FOLDER_NAME = "wait_folder_name"
	STATE_WAIT_INPUT_ROWS  = "wait_input_rows"
	STATE_WAIT_FOLDER_ID   = "wait_folder_id"
	STATE_IDLE             = "idle"
)

type dbRedis interface {
	AddRow()
	GetRow()
	DelRow()
	CheckUnfullfilledRows()
}

var userStates = make(map[int64]*UserState)
var stateMutex sync.RWMutex

func setState(chatID int64, state *UserState) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	state.CreatedAt = time.Now()
	userStates[chatID] = state
}

func getState(chatID int64) (*UserState, bool) {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	state, exists := userStates[chatID]
	return state, exists
}

func clearState(chatID int64) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	delete(userStates, chatID)
}

func welcomeMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info(fmt.Sprintf("User '%s' wrote '%s' for will start work", update.Message.Chat.Username, update.Message.Text))
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Привет!\nЧтобы посмотреть список доступных команд, введи /help",
	})
}
func helpMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info(fmt.Sprintf("User '%s' wrote '%s' for get help information", update.Message.Chat.Username, update.Message.Text))
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: `Список доступных команд:
/help - помощь по командам
/balance - посмотреть баланс
/get_folders_id - посмотреть существующие папки(Используется для разработки и модификации нашего бота)
/create_folder - создать папку с названием
/delete_folder - удалить папку
/create_folder - Создание новой папки (В разработке)
/create_task - создать задачу
/delete_task - удалить задачу или задачи
/restart - Перезапустить бота
/cancel - отменить действие и выйти в главное меню
Остальные команды в разработке 🙂
Связаться с разработчиком: @tatarkazawarka`,
	})
}

func checkBalance(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info(fmt.Sprintf("User '%s' wrote '%s' for check balance wallet", update.Message.Chat.Username, update.Message.Text))
	testObject := api.NewClient(os.Getenv("URL_UNU"), os.Getenv("UNU_API_TOKEN"))

	var firstObj api.UNUAPI = testObject
	balance := firstObj.Get_balance()
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Баланс вашего кошелька: %s", balance),
	})
}
func getFoldersId(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info(fmt.Sprintf("User '%s' wrote '%v' for get folder list id", update.Message.Chat.Username, update.Message.Text))
	testObject := api.NewClient(os.Getenv("URL_UNU"), os.Getenv("UNU_API_TOKEN"))

	var firstObj api.UNUAPI = testObject
	folder_list := firstObj.Get_folders()
	result_text := "Список папок:"
	for _, value := range folder_list {
		result_text += fmt.Sprintf("\nID: %s. Название: %s", value.ID.String(), value.Name)
	}
	result_text += "\nP.S Помните, что эта информация для вас бесполезна и используется только для разработчика 😉"
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Ваши папки: %s", result_text),
	})

}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {

	chatID := update.Message.Chat.ID
	state, exists := getState(chatID)

	if !exists {
		// Обычное сообщение, не связанное с состоянием
		return
	}

	// Проверяем время жизни состояния (максимум 5 минут)
	if time.Since(state.CreatedAt) > 5*time.Minute {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Время сессии истекло. Начните заново.",
		})
		clearState(chatID)
		return
	}

	switch state.State {
	case STATE_WAIT_FOLDER_NAME:
		handleFolderNameInput(ctx, b, update, state)
	case STATE_WAIT_FOLDER_ID:
		handleFolderIdInput(ctx, b, update, state)
	case STATE_WAIT_INPUT_ROWS:
		handleTaskRowInput(ctx, b, update, state)
	default:
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Неизвестное состояние.",
		})
		clearState(chatID)
	}

}
func handleFolderNameInput(ctx context.Context, b *bot.Bot, update *models.Update, state *UserState) {
	chatID := update.Message.Chat.ID
	folderName := strings.TrimSpace(update.Message.Text)

	if len(folderName) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Имя папки не может быть пустым. Введите имя еще раз:",
		})
		return
	}

	// Показываем что начали обработку
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   fmt.Sprintf("Создаю папку '%s'...", folderName),
	})

	// Создаем папку
	client := api.NewClient(os.Getenv("URL_UNU"), os.Getenv("UNU_API_TOKEN"))
	var clienObj api.UNUAPI = client
	folder_id, err := clienObj.Create_folder(folderName)

	if err != nil {
		slog.Error("Ошибка создания папки:", "ERROR:", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Ошибка при создании папки: %v", err),
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("✅ Папка '%s' успешно создана!\nID: %d", folderName, folder_id),
		})
	}

	clearState(chatID)
}
func createFolder(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info(fmt.Sprintf("User '%s' wrote '%s' for create folder", update.Message.Chat.Username, update.Message.Text))
	chatID := update.Message.Chat.ID

	setState(chatID, &UserState{
		State:   STATE_WAIT_FOLDER_NAME,
		Data:    make(map[string]interface{}),
		Command: "create_folder",
	})
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Пожалуйста, введите имя для папки которую хотим создать:",
	})

}
func deleteFolder(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info(fmt.Sprintf("User '%s' wrote '%s' for delete folder", update.Message.Chat.Username, update.Message.Text))
	chatID := update.Message.Chat.ID

	setState(chatID, &UserState{
		State:   STATE_WAIT_FOLDER_ID,
		Data:    make(map[string]interface{}),
		Command: "delete_folder",
	})
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Пожалуйста, введите ID для папки которую хотим удалить:",
	})

}
func handleFolderIdInput(ctx context.Context, b *bot.Bot, update *models.Update, state *UserState) {
	chatID := update.Message.Chat.ID
	folderId := strings.TrimSpace(update.Message.Text)

	if len(folderId) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "ID папки не может быть пустым. Введите ID еще раз:",
		})
		return
	}

	// Показываем что начали обработку
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   fmt.Sprintf("Удаляю папку '%s'...", folderId),
	})

	// Создаем папку
	client := api.NewClient(os.Getenv("URL_UNU"), os.Getenv("UNU_API_TOKEN"))
	var clienObj api.UNUAPI = client
	folderIdInt, err := strconv.Atoi(folderId)
	if err != nil {
		slog.Error("Не удалось преобразовать folderId в тип integer", "ERROR:", err)
	}
	_, err = clienObj.Delete_folder(folderIdInt)
	if err != nil {
		slog.Error("Ошибка создания папки:", "ERROR:", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Ошибка при удалении папки: %v", err),
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "✅ Папка успешно удалена!\n",
		})
	}

	clearState(chatID)
}

func createTask(ctx context.Context, b *bot.Bot, update *models.Update) {
	slog.Info(fmt.Sprintf("User '%s' wrote '%s' for create folder", update.Message.Chat.Username, update.Message.Text))
	chatID := update.Message.Chat.ID

	// TODO: Сделать здесь логику, чтобы при входе в данную функцию, сначала проверялась очередь.
	// Есть ли незавершенные задачи? Если есть, нужно ли обработать их в первую очередь или оставить на потом?

	// Проверили что задач нет, запрашиваем у клиента номера строк для выполнения
	setState(chatID, &UserState{
		State:   STATE_WAIT_INPUT_ROWS,
		Data:    make(map[string]interface{}),
		Command: "create_task",
	})
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Пожалуйста, введи номера строк для начала работы: Пример: 2-15(Не забывайте, что строка с номером 1, сервисная, на ней находятся названия колонок)",
	})

}
func handleTaskRowInput(ctx context.Context, b *bot.Bot, update *models.Update, state *UserState) {
	chatID := update.Message.Chat.ID
	folderName := strings.TrimSpace(update.Message.Text)

	if len(folderName) == 0 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Строки не могут быть пустым сообщением...",
		})
		return
	}
	rows := strings.Split(update.Message.Text, ":")
	beginRowString, endRowString := rows[0], rows[1]

	beginRowInt, err := strconv.Atoi(beginRowString)
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка конвертации значения для начальной строки... Просьба проверить корректность. Пользователь %s попытался ввёл: %s что привело к данной ошибке", update.Message.Chat.Username, update.Message.Text))
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Простите, вы ввели некорректное значение для начала работы. Пожалуйста, ориентируйтесь на пример который я вам показал",
		})
		clearState(chatID)
		return

	}
	endRowInt, err := strconv.Atoi(endRowString)
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка конвертации значения для начальной строки... Просьба проверить корректность. Пользователь %s попытался ввёл: %s что привело к данной ошибке", update.Message.Chat.Username, update.Message.Text))
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Простите, вы ввели некорректное значение для начала работы. Пожалуйста, ориентируйтесь на пример который я вам показал",
		})
		clearState(chatID)
		return
	}

	// Показываем что начали обработку
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "Создаю задачи...",
	})

	// Создаем папку
	client := api.NewClient(os.Getenv("URL_UNU"), os.Getenv("UNU_API_TOKEN"))
	var clienObj api.UNUAPI = client

	folder_id, err := clienObj.Add_task(beginRowInt, endRowInt)

	if err != nil {
		slog.Error("Ошибка создания задачи:", "ERROR:", err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("❌ Ошибка при создании: %v", err),
		})
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   fmt.Sprintf("✅  '%s' успешно создана!\nID: %d", folderName, folder_id),
		})
	}

	clearState(chatID)
}
