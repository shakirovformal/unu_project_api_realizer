package googlesheetreader

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func Reader() string {
	spreadsheetId := os.Getenv("SHEET_TABLE_ID")
	readRange := "BOT!A2:F2"
	ctx := context.Background()
	svc, err := sheets.NewService(ctx, option.WithCredentialsFile("creds.json"))
	if err != nil {
		log.Printf("\n Err is %v", err)
	}
	resp, err := svc.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	// // Обработка данных (в данном случае, печать)
	// if len(resp.Values) == 0 {
	// 	log.Println("No data found.")
	// } else {
	// 	log.Println("Data:")
	// 	for _, row := range resp.Values {
	// 		log.Printf("%s\n", row)
	// 	}
	// }
	fmt.Println("", fmt.Sprintf("type: %T. value: %v", resp.Values[0][1], resp.Values[0][1]))

	// TODO: сделать проверку, если описание больше чем 2300 символов в длину, то в дальнейшем будет ошибка,
	// надо предупредить клиента и постараться что-то с этим сделать
	return fmt.Sprint(resp.Values)
}
