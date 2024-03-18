package services

import "golang.org/x/crypto/bcrypt"

type PasswordService interface {
	CompareHashAndPassword(hashedPassword, password string) error
	GenerateHashedPassword(unhashedPassword string) (string, error)
}

type BcryptPasswordService struct{}

func NewBcryptPasswordService() *BcryptPasswordService {
	return &BcryptPasswordService{}
}

func (s *BcryptPasswordService) GenerateHashedPassword(unhashedPassword string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(unhashedPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err // Ensure you handle this error appropriately
	}
	return string(hashedPassword), nil
}

func (s *BcryptPasswordService) CompareHashAndPassword(passwordHashed, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(passwordHashed), []byte(password))
	if err != nil {
		return err
	}
	return nil
}
