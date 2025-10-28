package utils

import (
	"log/slog"
	"sort"
	"strconv"

	"github.com/shakirovformal/unu_project_api_realizer/pkg/models"
)

func ConverterUnfullfilledKeys(unFullFilledKeys []string) ([]int, error) {
	unFullFilledKeysInt := []int{}
	var err error
	var res int
	for _, value := range unFullFilledKeys {
		res, err = strconv.Atoi(value)
		if err != nil {
			slog.Error("Ошибка конвертации строки в число. Пожалуйста проверьте данные, которые отдала база данных")
			return nil, models.ErrorIncorrectData
		}
		unFullFilledKeysInt = append(unFullFilledKeysInt, res)
	}

	return SortByUp(unFullFilledKeysInt), nil

}

func SortByUp(keys []int) []int {
	sort.Ints(keys)
	return keys
}
