package models

import "errors"

var (
	// Errors
	ErrorGetValueFromDatabase = errors.New("Error get value from database")
	ErrorZeroValue            = errors.New("Getting zero value")
	ErrorUnmarshallJSON       = errors.New("Promblem with decoding from JSON")
	ErrorIncorrectData        = errors.New("Incorrect data")
	ErrorDatabase             = errors.New("Error with database")
	LongMessage               = errors.New("Long message. Length bigger 2300 symbols")
	ErrorMatchingSite         = errors.New("Error with matching choose site. Please check correct name")
	ErrorGoogleSheet          = errors.New("Error with getting value from google sheet")
	// Other
	GenderMale   = "мужской"
	GenderFemale = "женский"
)

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
