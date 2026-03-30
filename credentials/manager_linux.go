package credentials

import (
	"1c_cron_dump/models"
	"encoding/base64"
	"errors"
	"os"
	"strings"
)

func GetCredentials(infobase *models.Infobase) (error, string, string) {
	value, exists := os.LookupEnv(infobase.CredentialsVariable)

	if !exists {
		return errors.New("Credential not found"), "", ""
	}

	parts := strings.Split(value, ";")
	if len(parts) != 2 {
		return errors.New("Malformed credential string"), "", ""
	}

	decodedUsername, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return err, "", ""
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return err, "", ""
	}

	return nil, string(decodedUsername), string(decodedPassword)
}
