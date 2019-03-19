package cms

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore"
)

type Service interface {
	Put(id string, documentation []byte, apiSpec []byte, eventsSpec []byte) apperrors.AppError
	Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError)
	Remove(id string) apperrors.AppError
}

type service struct {
}

func NewService(repository assetstore.Repository) Service {
	return &service{}
}

func (s service) Put(id string, documentation []byte, apiSpec []byte, eventsSpec []byte) apperrors.AppError {
	return nil
}

func (s service) Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError) {
	return nil, nil, nil, nil
}

func (s service) Remove(id string) apperrors.AppError {
	return nil
}
