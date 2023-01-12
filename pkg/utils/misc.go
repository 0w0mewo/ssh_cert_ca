package utils

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"os/signal"
	"syscall"
)

func IsFileExist(fpath string) bool {
	_, err := os.Stat(fpath)
	return errors.Is(err, fs.ErrExist) || err == nil
}

func WaitForSignal() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)

	return ch
}

func RandomSha1Hex() string {
	sha1sum := sha1.New()

	randbytes := make([]byte, 24)
	_, err := rand.Read(randbytes)
	if err != nil {
		panic(err)
	}

	sha1sum.Write(randbytes)

	return hex.EncodeToString(sha1sum.Sum(nil))
}
