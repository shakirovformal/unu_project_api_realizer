package api

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	gsr "github.com/shakirovformal/unu_project_api_realizer/pkg/google-sheet-reader"
	"github.com/shakirovformal/unu_project_api_realizer/pkg/models"
	"google.golang.org/api/sheets/v4"
)

func getName(respData *sheets.ValueRange) (string, error) {

	// TODO: Название выстраивается за счёт данных:
	task_name := ""
	gettedDate := fmt.Sprint(respData.Values[0][5])
	publicationDate := normalizeData(gettedDate)
	// Получаем ссылку
	link := fmt.Sprint(respData.Values[0][1])
	ref, err := checkReferenceFromLink(link)
	if err != nil {
		slog.Error("Ошибка при попытке мэтчинга сайта по ссылке")
		return "", models.ErrorGoogleSheet
	}
	// Делаем проверку, что за ссылка. исходя из самой ссылки понимаем какой шаблон брать для использования
	// Получаем пол для выполнения задачи
	gender := checkGender(fmt.Sprint(respData.Values[0][2]))
	// Имя задачи: Дата + Шаблон из таблицы с учётом гендерности отзыва
	// Example: 27.10.2025 ОПУБЛИКОВАТЬ ГОТОВЫЙ отзыв мужской аккаунт

	task_name = fmt.Sprintf("%s %s %s отзыв", publicationDate, ref, gender)
	return task_name, nil
}

func checkGender(gender string) string {
	switch gender {
	case "м":
		return "мужской"
	case "ж":
		return "женский"
	}
	return ""
}

func normalizeData(dateString string) string {
	// Обработка пустых строк и пробелов
	dateString = strings.TrimSpace(dateString)
	if dateString == "" {
		return ""
	}

	// Сначала пробуем стандартные форматы через time.Parse
	if result := parseWithStandardLayouts(dateString); result != "" {
		return result
	}

	// Затем пробуем ручной парсинг с регулярными выражениями
	if result := parseWithRegex(dateString); result != "" {
		return result
	}

	// Если ничего не сработало, возвращаем исходную строку
	return ""
}

// parseWithStandardLayouts пытается распарсить дату с помощью стандартных layout'ов Go
func parseWithStandardLayouts(dateString string) string {
	layouts := []string{
		"02.01.2006",
		"2.1.2006",
		"02-01-2006",
		"2-1-2006",
		"02/01/2006",
		"2/1/2006",
		"2006-01-02",
		"2006-1-2",
		"02.01.06",
		"2.1.06",
		"02-01-06",
		"2-1-06",
		"02/01/06",
		"2/1/06",
		// Американские форматы
		"01/02/2006",
		"1/2/2006",
		"01-02-2006",
		"1-2-2006",
		"01.02.2006",
		"1.2.2006",
		"01/02/06",
		"1/2/06",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateString); err == nil {
			return t.Format("02.01.2006")
		}
	}

	return ""
}

// parseWithRegex парсит дату с помощью регулярных выражений
func parseWithRegex(dateString string) string {
	// Убираем пробелы вокруг разделителей для упрощения парсинга
	cleaned := regexp.MustCompile(`\s*([\.\-/])\s*`).ReplaceAllString(dateString, "$1")

	patterns := []struct {
		regex   *regexp.Regexp
		handler func([]string) (string, bool)
	}{
		// ДД.ММ.ГГГГ или ДД-ММ-ГГГГ или ДД/ММ/ГГГГ (европейский формат)
		{
			regex: regexp.MustCompile(`^(\d{1,2})[\.\-/](\d{1,2})[\.\-/](\d{4})$`),
			handler: func(matches []string) (string, bool) {
				day, month, year := matches[1], matches[2], matches[3]
				if isValidEuropeanDate(day, month, year) {
					return padZero(day) + "." + padZero(month) + "." + year, true
				}
				// Пробуем американский формат, если европейский не подошел
				if isValidAmericanDate(month, day, year) {
					return padZero(day) + "." + padZero(month) + "." + year, true
				}
				return "", false
			},
		},
		// ГГГГ-ММ-ДД или ГГГГ.ММ.ДД или ГГГГ/ММ/ДД (ISO формат)
		{
			regex: regexp.MustCompile(`^(\d{4})[\.\-/](\d{1,2})[\.\-/](\d{1,2})$`),
			handler: func(matches []string) (string, bool) {
				year, month, day := matches[1], matches[2], matches[3]
				if isValidDate(day, month, year) {
					return padZero(day) + "." + padZero(month) + "." + year, true
				}
				return "", false
			},
		},
		// ДД.ММ.ГГ или ДД-ММ-ГГ или ДД/ММ/ГГ (европейский с двухзначным годом)
		{
			regex: regexp.MustCompile(`^(\d{1,2})[\.\-/](\d{1,2})[\.\-/](\d{2})$`),
			handler: func(matches []string) (string, bool) {
				day, month, year := matches[1], matches[2], convertTwoDigitYear(matches[3])
				if isValidEuropeanDate(day, month, year) {
					return padZero(day) + "." + padZero(month) + "." + year, true
				}
				// Пробуем американский формат
				if isValidAmericanDate(month, day, year) {
					return padZero(day) + "." + padZero(month) + "." + year, true
				}
				return "", false
			},
		},
		// Американский формат ММ/ДД/ГГГГ или ММ-ДД-ГГГГ или ММ.ДД.ГГГГ
		{
			regex: regexp.MustCompile(`^(\d{1,2})[\.\-/](\d{1,2})[\.\-/](\d{4})$`),
			handler: func(matches []string) (string, bool) {
				month, day, year := matches[1], matches[2], matches[3]
				if isValidAmericanDate(month, day, year) {
					return padZero(day) + "." + padZero(month) + "." + year, true
				}
				return "", false
			},
		},
		// Американский формат с двухзначным годом ММ/ДД/ГГ
		{
			regex: regexp.MustCompile(`^(\d{1,2})[\.\-/](\d{1,2})[\.\-/](\d{2})$`),
			handler: func(matches []string) (string, bool) {
				month, day, year := matches[1], matches[2], convertTwoDigitYear(matches[3])
				if isValidAmericanDate(month, day, year) {
					return padZero(day) + "." + padZero(month) + "." + year, true
				}
				return "", false
			},
		},
	}

	for _, pattern := range patterns {
		if matches := pattern.regex.FindStringSubmatch(cleaned); matches != nil {
			if result, ok := pattern.handler(matches); ok {
				return result
			}
		}
	}

	return ""
}

// isValidEuropeanDate проверяет валидность даты в европейском формате (ДД.ММ.ГГГГ)
func isValidEuropeanDate(dayStr, monthStr, yearStr string) bool {
	return isValidDate(dayStr, monthStr, yearStr)
}

// isValidAmericanDate проверяет валидность даты в американском формате (ММ.ДД.ГГГГ)
func isValidAmericanDate(monthStr, dayStr, yearStr string) bool {
	// В американском формате первый элемент - месяц, второй - день
	return isValidDate(dayStr, monthStr, yearStr)
}

// isValidDate универсальная проверка валидности даты
func isValidDate(dayStr, monthStr, yearStr string) bool {
	day, err1 := strconv.Atoi(dayStr)
	month, err2 := strconv.Atoi(monthStr)
	year, err3 := strconv.Atoi(yearStr)

	if err1 != nil || err2 != nil || err3 != nil {
		return false
	}

	// Проверяем базовые границы
	if month < 1 || month > 12 || day < 1 || day > 31 || year < 1000 || year > 9999 {
		return false
	}

	// Проверяем количество дней в месяце
	daysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	// Учитываем високосный год для февраля
	if month == 2 {
		if (year%4 == 0 && year%100 != 0) || (year%400 == 0) {
			if day <= 29 {
				return true
			}
		}
	}

	if day <= daysInMonth[month-1] {
		return true
	}

	return false
}

// padZero добавляет ведущий ноль к числам меньше 10
func padZero(s string) string {
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

// convertTwoDigitYear преобразует двухзначный год в четырехзначный
func convertTwoDigitYear(year string) string {
	y, err := strconv.Atoi(year)
	if err != nil {
		return year
	}

	if y < 50 {
		return "20" + padZero(year)
	}
	return "19" + padZero(year)
}

func checkReferenceFromLink(link string) (string, error) {
	patternSlice := NewSiteMatcher()
	siteCell, err := patternSlice.GetCellForURL(link)
	if err != nil {
		return "", err
	}
	resp, err := gsr.ReaderFromCell(os.Getenv("SPREADSHEETID"), "REFERENCE", siteCell)
	textReference := fmt.Sprint(resp.Values[0][0])
	return textReference, nil
}

// SitePattern хранит шаблон URL и соответствующую ячейку
type SitePattern struct {
	Pattern *regexp.Regexp
	Cell    string
}

// SiteMatcher содержит все паттерны для сопоставления
type SiteMatcher struct {
	patterns []SitePattern
}

// NewSiteMatcher создает и инициализирует SiteMatcher с предопределенными паттернами
func NewSiteMatcher() *SiteMatcher {
	return &SiteMatcher{
		patterns: []SitePattern{
			{
				Pattern: regexp.MustCompile(`maps\.app\.goo\.gl`),
				Cell:    "A2",
			},
			{
				Pattern: regexp.MustCompile(`yandex\.(ru|com)/maps`),
				Cell:    "B2",
			},
			{
				Pattern: regexp.MustCompile(`otzovik\.com`),
				Cell:    "C2",
			},
			{
				Pattern: regexp.MustCompile(`irecommend\.ru`),
				Cell:    "D2",
			},
			{
				Pattern: regexp.MustCompile(`prodoctorov\.ru`),
				Cell:    "E2",
			},
			{
				Pattern: regexp.MustCompile(`sravni\.ru`),
				Cell:    "F2",
			},
		},
	}
}

// GetCellForURL возвращает ячейку для данного URL
func (sm *SiteMatcher) GetCellForURL(url string) (string, error) {
	for _, pattern := range sm.patterns {
		if pattern.Pattern.MatchString(url) {
			return pattern.Cell, nil
		}
	}
	return "", models.ErrorMatchingSite // или какое-то значение по умолчанию
}
