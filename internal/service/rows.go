package service

import (
	"context"

	pb "lowcode-mysql/api/lowcode/v1"
	"lowcode-mysql/internal/biz"
)

type RowsService struct {
	pb.UnimplementedRowsServer

	uc *biz.RowsUsecase
}

func NewRowsService(uc *biz.RowsUsecase) *RowsService {
	return &RowsService{uc: uc}
}

func (s *RowsService) RowsGet(ctx context.Context, req *pb.RowsGetRequest) (*pb.RowsGetReply, error) {
	return s.uc.RowsGet(ctx, req)
}
