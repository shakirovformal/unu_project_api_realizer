package database

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"testing"

	"github.com/shakirovformal/unu_project_api_realizer/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestAddRow(t *testing.T) {

	testNormalData := []*models.RowObject{
		// Нормальные данные
		models.NewRowObject(1, "убрир екб", "yandex.ru", 1, "Сделать отзыв", "25.10.2020"),
		// Специальные символы и HTML
		models.NewRowObject(3, "Тест", "https://site.com", 1, "Текст с <html> тегами", "15.07.2024"),
		// Очень длинные строки
		models.NewRowObject(4, "Очень длинное название проекта которое может превышать обычные лимиты",
			"https://very-long-domain-name-that-might-break-something.com",
			2,
			"Очень длинное описание текста которое может содержать много информации и должно корректно обрабатываться системой при сохранении и отображении в интерфейсе пользователя",
			"29.02.2024"),
		// Данные на разных языках
		models.NewRowObject(16, "Project Name", "site.com", 1, "English description", "01.01.2024"),
		models.NewRowObject(17, "项目名称", "site.com", 2, "中文描述", "01.01.2024"),
		models.NewRowObject(18, "プロジェクト名", "site.com", 1, "日本語の説明", "01.01.2024"),

		// NULL-эквиваленты (если ваша система их поддерживает)
		models.NewRowObject(19, "NULL", "NULL", 1, "NULL", "NULL"),
		// Невалидные URL
		models.NewRowObject(5, "Проект", "not-a-valid-url", 1, "Описание", "30.02.2023"), // невалидная дата
		models.NewRowObject(10, "Проект", "site.com", 1, "Описание", "invalid-date"),
		models.NewRowObject(11, "Проект", "site.com", 2, "Описание", "2024-13-45"), // невалидная дата
		// Юникод и эмодзи
		models.NewRowObject(12, "Проект 🚀", "site.com", 1, "Описание с эмодзи 👍 и Unicode 测试", "01.01.2024"),
		// Очень большой ID
		models.NewRowObject(999999, "Проект", "site.com", 1, "Описание", "01.01.2024"),
	}

	testDataWithError := []*models.RowObject{

		// Пустые строки
		models.NewRowObject(2, "", "https://example.com", 2, "", "01.01.2024"),

		// Нулевые значения
		models.NewRowObject(0, "Проект", "google.com", 0, "Описание", "31.12.2023"),

		// Крайние значения gender
		models.NewRowObject(6, "Проект", "site.com", -1, "Отрицательный гендер", "01.01.2024"),
		models.NewRowObject(7, "Проект", "site.com", 3, "Неизвестный гендер", "01.01.2024"),
		models.NewRowObject(8, "Проект", "site.com", 999, "Очень большой гендер", "01.01.2024"),

		// Нестандартные даты
		models.NewRowObject(9, "Проект", "site.com", 2, "Описание", ""), // пустая дата

		// Только пробелы
		models.NewRowObject(15, "   ", "site.com", 2, "   ", "01.01.2024"),

		// Отрицательный ID
		models.NewRowObject(-1, "Проект", "site.com", 2, "Описание", "01.01.2024"),
	}

	type Config struct {
		Addr     string
		Password string
		DB       int
	}
	var dbConfig Config = Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	var err error
	expErr := fmt.Sprint(models.ErrorIncorrectData)
	db := NewDB(dbConfig.Addr, dbConfig.Password, dbConfig.DB)
	rdb := db.Connect(db)
	for idx, value := range testNormalData {
		err := db.AddRow(context.TODO(), rdb, strconv.Itoa(idx), value)
		require.NoError(t, err, models.ErrorIncorrectData)
	}
	for idx, value := range testDataWithError {
		err = db.AddRow(context.TODO(), rdb, strconv.Itoa(idx), value)
		slog.Info("ERROR:", "ERROR:", err)
		require.EqualError(t, err, expErr, "Expected a specific error message")
	}

}

func TestGetRow(t *testing.T) {
	type Config struct {
		Addr     string
		Password string
		DB       int
	}
	var dbConfig Config = Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	var err error
	rowNumbersPositive := []string{"1", "2", "3"}
	rowNumbersNegative := []string{"0", "-1", "", "true"}
	db := NewDB(dbConfig.Addr, dbConfig.Password, dbConfig.DB)
	rdb := db.Connect(db)
	for _, v := range rowNumbersPositive {
		err = db.GetRow(context.TODO(), rdb, v)

		require.NoError(t, err)
	}
	for _, v := range rowNumbersNegative {
		err = db.GetRow(context.TODO(), rdb, v)
		require.Error(t, err)
	}

}

func TestDelRow(t *testing.T) {
	type Config struct {
		Addr     string
		Password string
		DB       int
	}
	var dbConfig Config = Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	var err error
	rowNumbersPositive := []string{"1", "2", "3"}
	rowNumbersNegative := []string{"0", "-1", "", "true"}
	db := NewDB(dbConfig.Addr, dbConfig.Password, dbConfig.DB)
	rdb := db.Connect(db)
	for _, v := range rowNumbersPositive {
		_, err = db.DelRow(context.TODO(), rdb, v)
		require.NoError(t, err)
	}
	for _, v := range rowNumbersNegative {
		_, err = db.DelRow(context.TODO(), rdb, v)
		require.Error(t, err)
	}

}
