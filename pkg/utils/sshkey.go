package utils

import (
	"golang.org/x/crypto/ssh"
)

func ParseSSHPublicKey(in []byte) (ssh.PublicKey, error) {
	key, _, _, _, err := ssh.ParseAuthorizedKey(in)

	return key, err

}

func IsSSHPublicKey(in []byte) error {
	_, _, _, _, err := ssh.ParseAuthorizedKey(in)

	return err
}
