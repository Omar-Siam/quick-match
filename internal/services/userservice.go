package services

import (
	"github.com/brianvoe/gofakeit/v6"
	"quick-match/internal/models"
)

func GenerateNewUser() models.UserDetails {
	gofakeit.Seed(0)

	unhashedPassword := generatePassword()
	ghp := BcryptPasswordService{}
	hashedPassword, _ := ghp.GenerateHashedPassword(unhashedPassword)

	return models.UserDetails{
		UserID:         gofakeit.UUID(),
		Email:          gofakeit.Email(),
		Password:       unhashedPassword,
		PasswordHashed: hashedPassword,
		Name:           gofakeit.Name(),
		Gender:         gofakeit.Gender(),
		Age:            gofakeit.Number(18, 50),
		Userlocation: models.Userlocation{
			Latitude:  gofakeit.Latitude(),
			Longitude: gofakeit.Longitude(),
		},
	}
}

func generatePassword() string {
	return gofakeit.Password(true, true, true, true, false, 12)
}
