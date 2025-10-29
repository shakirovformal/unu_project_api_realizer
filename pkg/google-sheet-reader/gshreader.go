package googlesheetreader

import (
	"context"
	"fmt"
	"log"

	"github.com/shakirovformal/unu_project_api_realizer/pkg/models"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Используем значения у результата resp.Values[0][index]:
// 0  - название проекта
// 1  - ссылка
// 2 - гендерный пол
// 3 - текст отзыва
// 5 - дата публикации
func Reader(spreadsheetId, rowNumber string) (*sheets.ValueRange, error) {

	readRange := fmt.Sprintf("BOT!A%s:F%s", rowNumber, rowNumber)

	ctx := context.Background()
	svc, err := sheets.NewService(ctx, option.WithCredentialsFile("creds.json"))
	if err != nil {
		log.Printf("\n Err is %v", err)
	}
	resp, err := svc.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	if len(fmt.Sprint(resp.Values[0][3])) > 2300 {
		// TODO: Проверка работает корректно. Нужно обработать кейс, что делать если длина комментария больше 2300 символов.
		return nil, models.LongMessage
	}

	// TODO: сделать проверку, если описание больше чем 2300 символов в длину, то в дальнейшем будет ошибка,
	// надо предупредить клиента и постараться что-то с этим сделать

	return resp, nil
}
