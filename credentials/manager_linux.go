package credentials

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
)

func GetCredentials(credId string) (error, string, string) {
	value, exists := os.LookupEnv(credId)

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
