// Code generated by goctl. DO NOT EDIT!
// Source: stock.proto

package stock

import (
	"context"

	"gozerodtm/stock-srv/pb"

	"github.com/tal-tech/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	DecuctReq  = pb.DecuctReq
	DeductResp = pb.DeductResp

	Stock interface {
		Deduct(ctx context.Context, in *DecuctReq, opts ...grpc.CallOption) (*DeductResp, error)
		DeductRollback(ctx context.Context, in *DecuctReq, opts ...grpc.CallOption) (*DeductResp, error)
	}

	defaultStock struct {
		cli zrpc.Client
	}
)

func NewStock(cli zrpc.Client) Stock {
	return &defaultStock{
		cli: cli,
	}
}

func (m *defaultStock) Deduct(ctx context.Context, in *DecuctReq, opts ...grpc.CallOption) (*DeductResp, error) {
	client := pb.NewStockClient(m.cli.Conn())
	return client.Deduct(ctx, in, opts...)
}

func (m *defaultStock) DeductRollback(ctx context.Context, in *DecuctReq, opts ...grpc.CallOption) (*DeductResp, error) {
	client := pb.NewStockClient(m.cli.Conn())
	return client.DeductRollback(ctx, in, opts...)
}
