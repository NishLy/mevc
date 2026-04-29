package database

import (
	"errors"

	apperror "github.com/NishLy/go-fiber-boilerplate/internal/error"
	"gorm.io/gorm"
)

func Wrap(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.NotFoundErr(err)
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return apperror.DuplicateErr(err)
	}

	return apperror.InternalErr(err)
}
