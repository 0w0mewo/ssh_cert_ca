package service

import (
	"encoding/base64"
	"time"

	"github.com/0w0mewo/ssh_cert_ca/internal/model"
	"github.com/0w0mewo/ssh_cert_ca/pkg/ca"
	"github.com/0w0mewo/ssh_cert_ca/pkg/repo/cert"
	"github.com/0w0mewo/ssh_cert_ca/pkg/utils"
	"golang.org/x/crypto/ssh"
)

type SSHCertCAService struct {
	certStore  cert.CertRepo
	kepair     *ca.CAKeyPairs
	role       model.RoleType
	revokeTask *utils.ScheduledTaskGroup
	cachedKRL  []byte
}

func NewSSHCertCAService(dbdriver, dsn string, privKeyFile, passparse string, role model.RoleType) (*SSHCertCAService, error) {
	kp, err := ca.LoadCAKeyPairs(privKeyFile, passparse)
	if err != nil {
		return nil, err
	}

	ret := &SSHCertCAService{
		certStore:  cert.NewCertRepo(dbdriver, dsn),
		kepair:     kp,
		role:       role,
		revokeTask: utils.NewScheduledTaskGroup("default"),
	}

	ret.regenerateRevokedList()

	ret.revokeTask.AddPerodical(1*time.Minute, func() error {
		return ret.taskRevokeExpiredCerts()
	})

	return ret, nil

}

// sign and store the new certificate
func (s *SSHCertCAService) Sign(pubkeyToSign ssh.PublicKey, keyid string, validPrincipals []string, ttl time.Duration) (c model.Cert, err error) {
	var isHost bool

	if model.CertTypeHost == s.role {
		isHost = true
	}

	c, err = s.kepair.Sign(pubkeyToSign, keyid, 0, validPrincipals, ttl, isHost)
	if err != nil {
		return
	}

	err = s.certStore.CreateCert(c)
	return

}

func (s *SSHCertCAService) PublicKeyAsAuthKeyStr() string {
	return s.kepair.PublicKeyAsAuthKeyStr()
}

func (s *SSHCertCAService) Revoke(keyid string) error {
	err := s.certStore.UpdateRevoke(keyid, true)
	if err != nil {
		return err
	}

	return s.regenerateRevokedList()
}

func (s *SSHCertCAService) regenerateRevokedList() (err error) {
	certs, err := s.certStore.GetRevokedCertIdsByRole(s.role)
	if err != nil {
		return
	}

	s.cachedKRL, err = s.kepair.GenerateRevokedList(certs...)
	if err != nil {
		return
	}

	return
}

func (s *SSHCertCAService) GetPresentRevokedListBase64() string {
	return base64.StdEncoding.EncodeToString(s.cachedKRL)
}

func (s *SSHCertCAService) Stop() error {
	s.revokeTask.WaitAndStop()
	return s.certStore.Close()
}

func (s *SSHCertCAService) taskRevokeExpiredCerts() error {
	certIds, err := s.certStore.GetExpiredCertIdsByRole(s.role)
	if err != nil {
		return err
	}

	for _, id := range certIds {
		err := s.Revoke(id)
		if err != nil {
			continue
		}
	}

	return nil

}
