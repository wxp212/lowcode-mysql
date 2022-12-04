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

func (s *RowsService) RowsCreate(ctx context.Context, req *pb.RowsCreateRequest) (*pb.RowsCreateReply, error) {
	return s.uc.RowsCreate(ctx, req)
}

func (s *RowsService) RowsGet(ctx context.Context, req *pb.RowsGetRequest) (*pb.RowsGetReply, error) {
	return s.uc.RowsGet(ctx, req)
}

func (s *RowsService) RowsUpdate(ctx context.Context, req *pb.RowsUpdateRequest) (*pb.RowsUpdateReply, error) {
	return s.uc.RowsUpdate(ctx, req)
}

func (s *RowsService) RowsDelete(ctx context.Context, req *pb.RowsDeleteRequest) (*pb.RowsDeleteReply, error) {
	return s.uc.RowsDelete(ctx, req)
}
