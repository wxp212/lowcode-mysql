package biz

import (
	"context"

	v1 "lowcode-mysql/api/lowcode/v1"
	"lowcode-mysql/internal/data"

	"github.com/go-kratos/kratos/v2/log"
)

type RowsUsecase struct {
	repo *data.RowsRepo
	log  *log.Helper
}

func NewRowsUsecase(repo *data.RowsRepo, logger log.Logger) *RowsUsecase {
	return &RowsUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *RowsUsecase) RowsCreate(ctx context.Context, req *v1.RowsCreateRequest) (*v1.RowsCreateReply, error) {
	return uc.repo.RowsCreate(ctx, req)
}

func (uc *RowsUsecase) RowsGet(ctx context.Context, req *v1.RowsGetRequest) (*v1.RowsGetReply, error) {
	return uc.repo.RowsGet(ctx, req)
}

func (uc *RowsUsecase) RowsUpdate(ctx context.Context, req *v1.RowsUpdateRequest) (*v1.RowsUpdateReply, error) {
	return uc.repo.RowsUpdate(ctx, req)
}

func (uc *RowsUsecase) RowsDelete(ctx context.Context, req *v1.RowsDeleteRequest) (*v1.RowsDeleteReply, error) {
	return uc.repo.RowsDelete(ctx, req)
}
