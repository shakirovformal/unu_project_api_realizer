package models


import "errors"

var ErrorGetValueFromDatabase = errors.New("Error get value from database")
var ErrorZeroValue = errors.New("Getting zero value")
var ErrorUnmarshallJSON = errors.New("Promblem with decoding from JSON")
var ErrorIncorrectData = errors.New("Incorrect data")
var ErrorDatabase = errors.New("Error with database")



type RowObject struct {
	UserId int `json:"userId"`
	Object struct {
		Project           string `json:"project"`
		Link              string `json:"link"`
		Gender            int    `json:"gender"` // 1 - женский, 2 - мужской
		TextDescription   string `json:"text_description"`
		DateOfPublication string `json:"date_of_publication"`
	} `json:"object"`
}

func NewRowObject(userId int, project, link string, gender int, textDescription, dateOfPublication string) *RowObject {
	return &RowObject{
		UserId: userId,
		Object: struct {
			Project           string `json:"project"`
			Link              string `json:"link"`
			Gender            int    `json:"gender"`
			TextDescription   string `json:"text_description"`
			DateOfPublication string `json:"date_of_publication"`
		}{
			Project:           project,
			Link:              link,
			Gender:            gender,
			TextDescription:   textDescription,
			DateOfPublication: dateOfPublication,
		},
	}
}

type Person struct {
	Name    string
	Details struct { // This is an inline anonymous struct
		Age  int
		City string
	}
}

// NewPerson is a factory function acting as a constructor
func NewPerson(name string, age int, city string) *Person {
	return &Person{
		Name: name,
		Details: struct { // Initialize the inline anonymous struct
			Age  int
			City string
		}{
			Age:  age,
			City: city,
		},
	}
}
