package model

import (
	"errors"
	"strings"
	"time"
)

const (
	CerTypeUser = iota
	CertTypeHost
)

type RoleType int

var ErrUnsupportedCertType = errors.New("unsupported cert type")

type Cert struct {
	KeyId      string    `json:"id" db:"keyid"`
	Type       RoleType  `json:"type" db:"type"`
	ValidStart time.Time `json:"valid_start" db:"valid_start"`
	ValidEnd   time.Time `json:"valid_end" db:"valid_end"`
	Content    string    `json:"cert_content" db:"content"`
	Revoked    bool      `json:"revoked" db:"revoked"`
}

func ParseCertType(certType string) (RoleType, error) {
	certType = strings.ToLower(certType)

	if strings.Contains(certType, "user") {
		return CerTypeUser, nil
	}

	if strings.Contains(certType, "host") {
		return CertTypeHost, nil
	}

	return -1, ErrUnsupportedCertType
}

func FormatType(role RoleType) string {
	switch role {
	case CerTypeUser:
		return "user"
	case CertTypeHost:
		return "host"
	}

	return "unsupported"
}
