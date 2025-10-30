package api

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/joho/godotenv"
	googlesheetreader "github.com/shakirovformal/unu_project_api_realizer/pkg/google-sheet-reader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/sheets/v4"
)

// func initEnv() {
// 	err := godotenv.Load("../env")
// 	if err != nil {
// 		log.Fatalf("Ошибка загрузки файла .env: %v", err)
// 	}
// }

type testCase struct {
	input       string
	expected    string
	description string
}

var validTestCases = []testCase{
	// Стандартные форматы с точками
	{"15.03.2023", "15.03.2023", "Стандартный ДД.ММ.ГГГГ"},
	{"05.12.2024", "05.12.2024", "С нулями ДД.ММ.ГГГГ"},
	{"5.3.2023", "05.03.2023", "Без нулей Д.М.ГГГГ"},
	{"1.12.2023", "01.12.2023", "День без нуля Д.ММ.ГГГГ"},
	{"15.1.2023", "15.01.2023", "Месяц без нуля ДД.М.ГГГГ"},

	// Форматы с дефисами
	{"15-03-2023", "15.03.2023", "Дефисы вместо точек"},
	{"5-3-2023", "05.03.2023", "Дефисы без нулей"},
	{"01-01-2024", "01.01.2024", "С нулями и дефисами"},

	// Форматы с косыми чертами
	{"15/03/2023", "15.03.2023", "Косые черты ДД/ММ/ГГГГ"},
	{"5/3/2023", "05.03.2023", "Косые черты без нулей"},
	{"01/12/2023", "01.12.2023", "С нулями и косыми чертами"},

	// ISO форматы
	{"2023-03-15", "15.03.2023", "ISO формат ГГГГ-ММ-ДД"},
	{"2024-12-05", "05.12.2024", "ISO с нулями"},

	// Двухзначный год
	{"15.03.23", "15.03.2023", "Двухзначный год 23 → 2023"},
	{"01.01.99", "01.01.1999", "Двухзначный год 99 → 1999"},
	{"15.03.50", "15.03.2050", "Двухзначный год 50 → 1950"},
	{"15-03-23", "15.03.2023", "Двухзначный год с дефисами"},
	{"15/03/49", "15.03.2049", "Двухзначный год с косыми чертами"},

	// Крайние значения дат
	{"31.12.2023", "31.12.2023", "Последний день года"},
	{"01.01.2023", "01.01.2023", "Первый день года"},
	{"29.02.2024", "29.02.2024", "29 февраля високосного года"},

	// С пробелами
	{"  15.03.2023  ", "15.03.2023", "С пробелами вокруг"},
	{"15 . 03 . 2023", "15.03.2023", "С пробелами вокруг разделителей"},
}

var invalidTestCases = []testCase{
	// Несуществующие даты
	{"32.01.2023", "", "Несуществующий день"},
	{"15.13.2023", "", "Несуществующий месяц"},
	{"29.02.2023", "", "29 февраля не високосного года"},
	{"31.04.2023", "", "31 апреля не существует"},

	// Некорректные форматы
	{"15.03", "", "Отсутствует год"},
	{"2023", "", "Только год"},
	{"15.03.20235", "", "Слишком длинный год"},
	{"15.03.2", "", "Слишком короткий год"},

	// Смешанные разделители
	{"15.03-2023", "15.03.2023", "Смешанные разделители"},
	{"15/03.2023", "15.03.2023", "Смешанные разделители 2"},

	// Текст и специальные символы
	{"abc.def.ghij", "", "Текст вместо цифр"},
	{"15.03.2023!", "", "Спецсимвол в конце"},
	{"", "", "Пустая строка"},
	{"  ", "", "Только пробелы"},

	// Нестандартные форматы
	{"15 Mar 2023", "", "Текстовый месяц"},
	{"2023/03/15", "15.03.2023", "Формат ГГГГ/ММ/ДД"},
	{"03/15/2023", "15.03.2023", "Американский формат ММ/ДД/ГГГГ"},
}

func TestNormalizeDate(t *testing.T) {
	for _, value := range validTestCases {
		fmt.Println(value.input)
		result := normalizeData(value.input)
		if !assert.Equal(t, value.expected, result) {
			fmt.Printf("Description: %s\n Expected: %s\n Getted: %s\n", value.description, value.expected, result)
		}
	}
	fmt.Println("===========")
	for _, value := range invalidTestCases {
		fmt.Println(value.input)
		result := normalizeData(value.input)
		if !assert.Equal(t, value.expected, result) {
			fmt.Printf("Description: %s\n Expected: %s\n Getted: %s\n", value.description, value.expected, result)
		}
	}

}

type useCasesStructGetName struct {
	resp   *sheets.ValueRange
	expRes string
}

func NewStructGetName(resp *sheets.ValueRange, expres string) *useCasesStructGetName {
	return &useCasesStructGetName{
		resp:   resp,
		expRes: expres,
	}
}

type sheetValue struct {
	resp      *sheets.ValueRange
	sheetName string
	row       string
	err       error
}

func NewSheetValue(spreadsheetId string, sheetName string, row string) *sheetValue {
	response, err := googlesheetreader.Reader(spreadsheetId, sheetName, row)
	if err != nil {
		slog.Warn("WARN with google sheet test", "WARN", err)
	}
	return &sheetValue{
		resp: response,
	}
}

func TestGetName(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Ошибка загрузки файла .env: %v", err)
	}

	spreadsheetId := os.Getenv("SPREADSHEETID")

	respStruct := NewSheetValue(spreadsheetId, "BOT", "3")
	useCase := []*useCasesStructGetName{
		{
			resp:   respStruct.resp,
			expRes: "12.05.2025 ОПУБЛИКОВАТЬ ГОТОВЫЙ женский отзыв",
		},
	}

	for _, value := range useCase {
		result, err := getName(value.resp)
		require.NoError(t, err)
		if assert.Equal(t, value.expRes, result) {
			fmt.Println("EXPECTED:", value.expRes, "\nGOT:", result)
		}
	}

}
