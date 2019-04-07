package envsecrets

import "github.com/joho/godotenv"

func LoadEnvfile(file string) (map[string]string, error) {
	return godotenv.Read()
}
