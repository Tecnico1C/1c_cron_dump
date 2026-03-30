package credentials

import (
	"syscall"

	"github.com/danieljoos/wincred"
)

func GetCredentials(credId string) (error, string, string) {
	cred, err := wincred.GetGenericCredential(credId)
	if err != nil {
		return err, "", ""
	}
	username := cred.UserName
	u16 := make([]uint16, len(cred.CredentialBlob)/2)

	for i := range len(u16) {
		u16[i] = uint16(cred.CredentialBlob[i*2]) |
			uint16(cred.CredentialBlob[i*2+1])<<8
	}
	password := syscall.UTF16ToString(u16)

	return nil, username, password

}
