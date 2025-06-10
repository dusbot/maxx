package crack

import (
	"context"
	"errors"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioCracker struct {
	CrackBase
}

func (r *MinioCracker) Ping() (succ bool, err error) {
	return false, errors.ErrUnsupported
}

func (m *MinioCracker) Crack() (succ bool, err error) {
	ctx := context.Background()
	minioClient, err := minio.New(m.Target, &minio.Options{
		Creds: credentials.NewStaticV4(m.User, m.Pass, ""),
	})
	if err != nil {
		return false, err
	}
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		return false, err
	}
	succ = true
	return
}
