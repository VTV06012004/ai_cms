package domain

import (
	"errors"
	"time"
)

var ErrMediaNotFound = errors.New("không tìm thấy ảnh")

type MediaFile struct {
	Name       string
	Size       int64
	ModifiedAt time.Time
}
