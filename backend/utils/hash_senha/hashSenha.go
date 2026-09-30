package hashsenha

import "golang.org/x/crypto/bcrypt"

func Encriptar(senha string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), err
}

func CompararSenha(hash, senha string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
}
