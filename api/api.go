package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Client struct {
	client_url   string
	client_token string
}

func NewClient(input_url, input_token string) *Client {
	return &Client{
		client_url:   input_url,
		client_token: input_token,
	}
}

func (c Client) post(action string, params map[string]interface{}) string {
	formData := url.Values{
		"api_key": {c.client_token},
		"action":  {action}}
	for key, value := range params {
		switch v := value.(type) {
		case string:
			formData.Add(key, v)
		case int:
			formData.Add(key, strconv.Itoa(v))
		case int64:
			formData.Add(key, strconv.FormatInt(v, 10))
		case float64:
			formData.Add(key, strconv.FormatFloat(v, 'f', -1, 64))
		case bool:
			formData.Add(key, strconv.FormatBool(v))
		default:
			// Пробуем преобразовать в строку через fmt
			formData.Add(key, fmt.Sprintf("%v", v))
		}
	}
	resp, err := http.PostForm(os.Getenv("URL_UNU"), formData)
	if err != nil {
		log.Fatalf("Ошибка %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка %v", err)
	}
	bodyString := string(body)
	return bodyString
}

type UNUAPI interface {
	Get_balance() string
	Get_folders() []struct {
		ID   json.Number `json:"id"`
		Name string      `json:"name"`
	}
	Create_folder(folder_name string) (int64, error)
	Delete_folder(folder_id int) (bool, error)
	// Move_task()
	// Get_tasks()
	// Get_reports()
	// Approve_report()
	// Reject_report()
	// Get_expenses()
	Add_task(beginRow, endRow int) (int, error)
	// Del_task()
	// Task_limit_add()
	// Edit_task()
	// Get_tariffs()
	// Task_pause()
	// Task_play()

}

func (c *Client) Get_balance() string {

	type Response struct {
		Success bool    `json:"success"`
		Errors  string  `json:"errors"`
		Balance float64 `json:"balance"`
		Freeze  float64 `json:"freeze"`
	}

	slog.Info("goes to API for get balance wallet")
	bytesRes := c.post("get_balance", nil)

	var response Response
	err := json.Unmarshal([]byte(bytesRes), &response)
	if err != nil {
		slog.Warn("Ошибка парсинга JSON:", "ERROR:", err)
	}

	return fmt.Sprint(response.Balance)
}

func (c *Client) Get_folders() []struct {
	ID   json.Number `json:"id"`
	Name string      `json:"name"`
} {
	type Response struct {
		Success bool   `json:"success"`
		Errors  string `json:"errors"`
		Folders []struct {
			ID   json.Number `json:"id"`
			Name string      `json:"name"`
		} `json:"folders"`
	}

	slog.Info("goes to API for get folder list id`s")
	bytesRes := c.post("get_folders", nil)
	fmt.Println(bytesRes)
	var response Response
	err := json.Unmarshal([]byte(bytesRes), &response)
	if err != nil {
		slog.Warn("Ошибка парсинга JSON:", "ERROR:", err)
	}
	slog.Info("Success get folders")

	return response.Folders
}

func (c *Client) Create_folder(folder_name string) (int64, error) {
	action_value := make(map[string]interface{})
	action_value["name"] = folder_name
	slog.Info(fmt.Sprintf("Creating folder with name %s", folder_name))
	type Response struct {
		Success   bool        `json:"success"`
		Errors    string      `json:"errors"`
		Folder_ID json.Number `json:"folder_id"`
		Freeze    float64     `json:"freeze"`
	}

	bytesRes := c.post("create_folder", action_value)
	slog.Debug("We get result for creating:", "GETIING:", bytesRes)
	var response Response
	err := json.Unmarshal([]byte(bytesRes), &response)
	if err != nil {
		slog.Warn("Ошибка парсинга JSON:", "ERROR:", err)
		return 0, err
	}
	slog.Info("Success create folders")
	folder_id, err := response.Folder_ID.Int64()
	if err != nil {
		slog.Warn("Error for")
		return 0, err
	}

	return folder_id, nil

}
func (c *Client) Delete_folder(folder_id int) (bool, error) {
	action_value := make(map[string]interface{})
	action_value["folder_id"] = folder_id
	slog.Info(fmt.Sprintf("Deleting folder with name %d", folder_id))
	type Response struct {
		Success bool   `json:"success"`
		Errors  string `json:"errors"`
	}

	bytesRes := c.post("del_folder", action_value)
	fmt.Println(bytesRes)
	var response Response

	err := json.Unmarshal([]byte(bytesRes), &response)
	if err != nil {
		slog.Warn("Ошибка парсинга JSON:", "ERROR:", err)
		return false, err
	}
	slog.Info("Success delete folder")
	if response.Success == false {
		slog.Error("Ошибка при удалении")
		return false, err
	}

	return true, nil
}

//     name (text) – название задачи
//     descr (text) – текст задания
//     link (text) – URL, необходимый для выполнения задания (необязательный параметр)
//     need_for_report (text) – что должен предоставить исполнитель для отчёта по задаче
//     price (float) – стоимость одного выполнения задачи в рублях
//     tarif_id (int) – идентификатор тарифа
//     folder_id (int) – идентификатор папки, в которую нужно поместить задачу
//     need_screen (boolean) – если в задании исполнителю нужно прикерпить скриншот, нужно передать 1 (необязательный параметр)
//     time_for_work (int) – сколько часов дать исполнителю для работы, от 2 до 168 (необязательный параметр)
//     time_for_check (int) – сколько часов вам нужно для проверки задания, от 10 до 168 (необязательный параметр)
//     targeting_gender (int) – параметр таргетинга: пол. 1 – женский, 2 – мужской (необязательный параметр)
//     targeting_geo_country_id (int) – параметр геотаргетинга: ID страны (необязательный параметр)

// Выходные данные

// task_id (int) – идентификатор созданной задачи

func (c *Client) Add_task(beginRow, endRow int) (int, error) {
	//"""Обработка в функции идёт только 1 строки"""

	//TODO: Пойти в таблицу и получить строку

	//TODO: Получить имя для задачи

	//TODO: Получить описание задания

	//TODO: Получить ссылку для задания

	//TODO: Получить данные: что нужно для выполнения задания

	//TODO: получить стоимость задания

	//TODO: понять, какой тариф выбрать

	//TODO: понять в какую папку сохранить задание

	//TODO: Необходимость скриншота (по умолчанию всегда True)

	//TODO: Время на выполнение 72 часа

	//TODO: Время на проверку 120 часов

	//TODO: Получить значение гендерного пола для задания

	//TODO: Выбрать страну: Россия для задания

	//TODO: Добавить в базу значение строки

	//TODO: Отправить запрос на API

	//TODO: Если ответ успешный, удалить из базы

	return 0, nil
}
