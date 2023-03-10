package cert

import (
	"github.com/0w0mewo/ssh_cert_ca/internal/model"
)

type CertRepo interface {
	CreateCert(cert model.Cert) error
	UpdateRevoke(certId string, revoked bool) error
	GetCertsByRole(role model.RoleType) ([]*model.Cert, error)
	// GetCertsById(id string) ([]*model.Cert, error)
	GetRevokedCertIdsByRole(role model.RoleType) ([]string, error)
	GetExpiredCertIdsByRole(role model.RoleType) ([]string, error)
	Close() error
}

func NewCertRepo(driver, dsn string) CertRepo {
	switch driver {
	case "memory":
		return NewMemStore()
	case "sqlite3":
		return NewSqlRepo("sqlite", dsn)
	case "mysql":
		return NewSqlRepo("mysql", dsn)
	}

	return NewMemStore()
}
