package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/shakirovformal/unu_project_api_realizer/pkg/models"
)

var ctx = context.Background()

type Db struct {
	Addr     string
	Password string
	DB       int
}

func NewDB(addr, password string, db int) *Db {
	return &Db{
		Addr:     addr,
		Password: password,
		DB:       db,
	}
}

func (db *Db) Connect(database *Db) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     database.Addr,
		Password: database.Password,
		DB:       database.DB,
	})
	return rdb
}

func (db *Db) AddRow(ctx context.Context, rdb *redis.Client, rowNumber string, rowObject *models.RowObject) error {

	err := validateRowObject(rowNumber, rowObject)
	if err != nil {
		return err
	}

	dbObjPrepared, err := json.Marshal(rowObject)
	if err != nil {
		slog.Error("Ошибка маршаллинга структуры для сохранения в БД в формате JSON", "ERROR", err)
		return err
	}
	err = rdb.Set(ctx, rowNumber, string(dbObjPrepared), 0).Err()
	if err != nil {
		slog.Error("Ошибка создания ключа в базе данных", "ERROR", err)

		return models.ErrorGetValueFromDatabase
	}
	return nil
}

func (db *Db) GetRow(ctx context.Context, rdb *redis.Client, rowNumber string) error {
	err := validateRowNumber(rowNumber)
	if err != nil {
		return models.ErrorIncorrectData
	}
	gettingRes, err := rdb.Get(ctx, rowNumber).Result()
	if err != nil {
		slog.Error(fmt.Sprintf("Ошибка получения значения с ключом %s", rowNumber), "ERROR", err)
		return models.ErrorDatabase
	}
	var unmarshalStruct models.RowObject

	err = json.Unmarshal([]byte(gettingRes), &unmarshalStruct)
	if err != nil {
		slog.Error("Проблема размаршалливания JSON в структуру", "ERROR", err)
		return models.ErrorUnmarshallJSON
	}
	fmt.Println("Struct:", unmarshalStruct.Object.Project)
	return nil
}

func (db *Db) DelRow(ctx context.Context, rdb *redis.Client, rowNumber string) (int64, error) {
	err := validateRowNumber(rowNumber)
	if err != nil {
		return 0, models.ErrorIncorrectData
	}
	res, err := rdb.Del(ctx, rowNumber).Result()
	if err != nil {
		return 0, models.ErrorDatabase
	}
	return res, nil
}

func (db *Db) CheckUnfullfilledRows(ctx context.Context, rdb *redis.Client) ([]string, error) {
	sliceKeys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		return nil, models.ErrorDatabase
	}
	fmt.Println(sliceKeys)
	return sliceKeys, nil
}

func validateRowObject(rowNumber string, obj *models.RowObject) error {

	if len(rowNumber) == 0 {
		slog.Error("Не получен номер строки для добавления в базу данных")
		return models.ErrorIncorrectData
	}
	if obj == nil {
		slog.Error("Не получен объект для добавления в базу данных")
		return models.ErrorIncorrectData
	}
	if obj.UserId <= 0 {
		slog.Error("User ID не может быть нулевым")
		return models.ErrorIncorrectData
	}
	if obj.Object.Project == "" || len(strings.ReplaceAll(obj.Object.Project, " ", "")) == 0 {
		slog.Error("Не получено название проекта для добавления в базу данных")
		return models.ErrorIncorrectData
	}
	if obj.Object.Link == "" {
		slog.Error("Не получена ссылка для добавления в базу данных")
		return models.ErrorIncorrectData
	}
	if obj.Object.TextDescription == "" {
		slog.Error("Не получено описание работы для добавления в базу данных")
		return models.ErrorIncorrectData
	}
	if obj.Object.DateOfPublication == "" {
		slog.Error("Не получена дата для публикации для добавления в базу данных")
		return models.ErrorIncorrectData
	}
	if obj.Object.Gender <= 0 || obj.Object.Gender > 2 {
		slog.Error("Не валидное значение гендерного пола для добавления в базу данных")
		return models.ErrorIncorrectData
	}
	return nil
}
func validateRowNumber(rowNumber string) error {
	rowNumberInt, err := strconv.Atoi(rowNumber)
	if err != nil {
		return models.ErrorIncorrectData
	}
	if rowNumber == "" || len(strings.ReplaceAll(rowNumber, " ", "")) == 0 {
		return models.ErrorIncorrectData
	}
	if rowNumberInt <= 0 {
		return models.ErrorIncorrectData
	}
	return nil
}
