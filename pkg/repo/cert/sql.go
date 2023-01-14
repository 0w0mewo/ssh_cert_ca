package cert

import (
	"time"

	"github.com/0w0mewo/ssh_cert_ca/internal/model"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type stmts struct {
	createCert               *sqlx.Stmt
	getAllCertsByRole        *sqlx.Stmt
	getAllRevokedCertsByRole *sqlx.Stmt
	getAllExpiredCertsByRole *sqlx.Stmt
	updateRevoked            *sqlx.Stmt
}

type SqlStore struct {
	preparedStmts *stmts
	db            *sqlx.DB
}

func prepareStmts(db *sqlx.DB) (stmt *stmts, err error) {
	stmt = &stmts{}

	stmt.createCert, err = db.Preparex("INSERT INTO certs (keyid, type, valid_start, valid_end, content, revoked) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return
	}

	stmt.getAllCertsByRole, err = db.Preparex("SELECT * FROM certs WHERE type = ? AND revoked != 1")
	if err != nil {
		return
	}

	stmt.getAllRevokedCertsByRole, err = db.Preparex("SELECT keyid FROM certs WHERE type = ? AND revoked = 1")
	if err != nil {
		return
	}

	stmt.getAllExpiredCertsByRole, err = db.Preparex("SELECT keyid FROM certs WHERE type = ? AND revoked = 0 AND valid_end < ?")
	if err != nil {
		return
	}

	stmt.updateRevoked, err = db.Preparex("UPDATE certs SET revoked = ? WHERE keyid = ?")
	if err != nil {
		return
	}

	return

}

func NewSqlRepo(sqldriver, dsn string) *SqlStore {
	db, err := sqlx.Connect(sqldriver, dsn)
	if err != nil {
		panic(err)
	}

	stmt, err := prepareStmts(db)
	if err != nil {
		panic(err)
	}

	ret := &SqlStore{
		db:            db,
		preparedStmts: stmt,
	}

	err = ret.migration()
	if err != nil {
		panic(err)
	}

	return ret

}

func (ss *SqlStore) migration() error {
	// make sure table exist
	_, err := ss.db.Exec("CREATE TABLE IF NOT EXISTS certs (keyid VARCHAR(50) PRIMARY KEY, type TINYINT, valid_start DATETIME, valid_end DATETIME, content TEXT, revoked BOOLEAN)")
	if err != nil {
		return err
	}

	// make index
	_, err = ss.db.Exec("CREATE INDEX IF NOT EXISTS idx_role_revoke ON certs(type, revoked)")
	if err != nil {
		return err
	}

	return nil
}

func (ss *SqlStore) CreateCert(cert model.Cert) error {
	_, err := ss.preparedStmts.createCert.Exec(cert.KeyId, cert.Type, cert.ValidStart, cert.ValidEnd, cert.Content, cert.Revoked)

	return err
}

func (ss *SqlStore) UpdateRevoke(certId string, revoked bool) error {
	_, err := ss.preparedStmts.updateRevoked.Exec(revoked, certId)

	return err
}

func (ss *SqlStore) GetCertsByRole(role model.RoleType) ([]*model.Cert, error) {
	res := make([]*model.Cert, 0)
	err := ss.preparedStmts.getAllCertsByRole.Select(&res, role)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ss *SqlStore) GetRevokedCertIdsByRole(role model.RoleType) ([]string, error) {
	res := make([]string, 0)
	err := ss.preparedStmts.getAllRevokedCertsByRole.Select(&res, role)
	if err != nil {
		return nil, err
	}
	
	return res, nil
}

func (ss *SqlStore) GetExpiredCertIdsByRole(role model.RoleType) ([]string, error) {
	res := make([]string, 0)
	err := ss.preparedStmts.getAllExpiredCertsByRole.Select(&res, role, time.Now())
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ss *SqlStore) Close() error {
	return ss.db.Close()
}
