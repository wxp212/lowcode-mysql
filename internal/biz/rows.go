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

func (uc *RowsUsecase) RowsGet(ctx context.Context, req *v1.RowsGetRequest) (*v1.RowsGetReply, error) {
	return uc.repo.RowsGet(ctx, req)
}
