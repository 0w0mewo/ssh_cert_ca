package cert

import (
	"sync"
	"time"

	"github.com/0w0mewo/ssh_cert_ca/internal/model"
	"github.com/0w0mewo/ssh_cert_ca/pkg/repo"
)

type MemStore struct {
	store map[string]model.Cert
	lock  *sync.Mutex
}

func NewMemStore() *MemStore {
	return &MemStore{
		store: make(map[string]model.Cert),
		lock:  &sync.Mutex{},
	}
}

func (m *MemStore) CreateCert(cert model.Cert) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.store[cert.KeyId] = cert

	return nil
}

func (m *MemStore) UpdateRevoke(certId string, revoked bool) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	c, exist := m.store[certId]
	if !exist {
		return repo.ErrNotExist
	}

	c.Revoked = revoked
	m.store[certId] = c

	return nil
}

func (m *MemStore) GetCertsByRole(role model.RoleType) ([]*model.Cert, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	res := make([]*model.Cert, 0)

	for _, c := range m.store {
		if c.Type == role {
			res = append(res, &c)
		}
	}

	return res, nil
}

func (m *MemStore) GetRevokedCertByRole(role model.RoleType) ([]*model.Cert, error) {
	certs, err := m.GetCertsByRole(role)
	if err != nil {
		return nil, err
	}

	res := make([]*model.Cert, 0)

	for _, c := range certs {
		if c.Revoked {
			res = append(res, c)
		}
	}

	return res, nil
}

func (m *MemStore) GetExpiredCertsByRole(role model.RoleType) ([]*model.Cert, error) {
	certs, err := m.GetCertsByRole(role)
	if err != nil {
		return nil, err
	}

	res := make([]*model.Cert, 0)

	for _, c := range certs {
		if !c.Revoked && c.ValidEnd.After(time.Now()){
			res = append(res, c)
		}
	}

	return res, nil
}

func (m *MemStore) Close() error {
	return nil
}