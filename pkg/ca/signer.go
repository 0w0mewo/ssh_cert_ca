package ca

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/0w0mewo/ssh_cert_ca/internal/model"
	"github.com/0w0mewo/ssh_cert_ca/pkg/utils"
	"github.com/stripe/krl"
	"golang.org/x/crypto/ssh"
)

type SignerFunc func(pubkeyToSign ssh.PublicKey, keyid string, serial uint64, hostnames []string, ttl time.Duration) (c model.Cert, err error)

type CAKeyPairs struct {
	pubkey  ssh.PublicKey
	privkey ssh.Signer
}

// load ssh CA keypairs from file
func LoadCAKeyPairs(privateKeyFile, passparse string) (kp *CAKeyPairs, err error) {
	// generate keypair if it's not exist
	if !utils.IsFileExist(privateKeyFile) {
		err = generateKeyECPair(privateKeyFile, passparse)
		if err != nil {
			return
		}
	}

	// private key
	privkeyBytes, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return
	}

	var privkey ssh.Signer
	if passparse == "" {
		privkey, err = ssh.ParsePrivateKey(privkeyBytes)
		if err != nil {
			return
		}
	} else {
		privkey, err = ssh.ParsePrivateKeyWithPassphrase(privkeyBytes, []byte(passparse))
		if err != nil {
			return
		}

	}

	kp = &CAKeyPairs{
		pubkey:  privkey.PublicKey(),
		privkey: privkey,
	}

	return

}

func generateKeyECPair(privateKeyFile, passparse string) error {
	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return err
	}

	keyfile, err := os.OpenFile(privateKeyFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer keyfile.Close()

	keybytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}

	return pem.Encode(keyfile, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keybytes,
	})
}

func (ckp *CAKeyPairs) PublicKeyAsAuthKeyStr() string {
	return string(bytes.Trim(ssh.MarshalAuthorizedKey(ckp.pubkey), "\n"))

}

func (ckp *CAKeyPairs) Sign(pubkeyToSign ssh.PublicKey, keyid string, serial uint64, validPrincipals []string, ttl time.Duration, isHost bool) (c model.Cert, err error) {
	nonce := make([]byte, 32)
	_, err = rand.Read(nonce)
	if err != nil {
		return
	}

	var certType uint32
	principals := make([]string, 0)

	if isHost {
		certType = ssh.HostCert
		c.Type = model.CertTypeHost
		principals = append(principals, validPrincipals...)
	} else {
		certType = ssh.UserCert
		c.Type = model.CerTypeUser
		principals = append(principals, validPrincipals[0])
	}

	c.ValidStart = time.Now()
	c.ValidEnd = time.Now().Add(ttl)
	c.KeyId = keyid

	cert := &ssh.Certificate{
		Nonce:           nonce,
		Key:             pubkeyToSign,
		Serial:          serial,
		CertType:        certType,
		KeyId:           keyid,
		ValidPrincipals: principals,
		ValidAfter:      uint64(c.ValidStart.Unix()),
		ValidBefore:     uint64(c.ValidEnd.Unix()),
		Permissions: ssh.Permissions{
			Extensions: map[string]string{
				"permit-X11-forwarding":   "",
				"permit-agent-forwarding": "",
				"permit-port-forwarding":  "",
				"permit-pty":              "",
				"permit-user-rc":          "",
			},
		},
	}

	err = cert.SignCert(rand.Reader, ckp.privkey)
	if err != nil {
		return
	}

	c.Content = fmt.Sprintf("%s %s", cert.Type(), base64.StdEncoding.EncodeToString(cert.Marshal()))

	return
}

func (ckp *CAKeyPairs) GenerateRevokedList(certs ...*model.Cert) ([]byte, error) {
	reovkedCerts := &krl.KRLCertificateSection{
		CA: ckp.pubkey,
	}

	ids := krl.KRLCertificateKeyID{}
	for _, cert := range certs {
		if cert.Revoked {
			ids = append(ids, cert.KeyId)
		}
	}

	reovkedCerts.Sections = append(reovkedCerts.Sections, &ids)
	k := &krl.KRL{
		Sections: []krl.KRLSection{reovkedCerts},
	}

	return k.Marshal(rand.Reader, ckp.privkey)

}
