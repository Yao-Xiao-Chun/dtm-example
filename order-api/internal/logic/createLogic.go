package logic

import (
	"context"
	"fmt"
	"github.com/yedf/dtmcli/dtmimp"
	"gozerodtm/order-srv/order"
	"gozerodtm/stock-srv/stock"
	"net/http"

	"gozerodtm/order-api/internal/svc"
	"gozerodtm/order-api/internal/types"

	"github.com/tal-tech/go-zero/core/logx"
	"github.com/yedf/dtmgrpc"
)

// dtm已经通过配置，注册到下面这个地址，因此在dtmgrpc中使用该地址
var dtmServer = "etcd://127.0.0.1:2379/dtmservice"

//var dtmServer = "consul://127.0.0.1:8500/dtmservice"

type CreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) CreateLogic {
	return CreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Create 分布式事务 创建订单 减少库存
func (l *CreateLogic) Create(req types.QuickCreateReq, r *http.Request) (*types.QuickCreateResp, error) {

	orderRpcBusiServer, err := l.svcCtx.Config.OrderRpcConf.BuildTarget() //获取order rpc地址

	if err != nil {
		return nil, fmt.Errorf("下单异常超时")
	}
	stockRpcBusiServer, err := l.svcCtx.Config.StockRpcConf.BuildTarget() // 获取stock rpc地址
	if err != nil {
		return nil, fmt.Errorf("下单异常超时")
	}
	fmt.Println(orderRpcBusiServer, stockRpcBusiServer)
	//创建订单请求值
	createOrderReq := &order.CreateReq{UserId: req.UserId, GoodsId: req.GoodsId, Num: req.Num}

	//库存扣减
	deductReq := &stock.DecuctReq{GoodsId: req.GoodsId, Num: req.Num}

	//这里只举了saga例子，tcc等其他例子基本没啥区别具体可以看dtm官网

	gid := dtmgrpc.MustGenGid(dtmServer) //生成 全局事务id

	//开启分布式事务
	saga := dtmgrpc.NewSagaGrpc(dtmServer, gid).
		Add(orderRpcBusiServer+"/pb.order/create", orderRpcBusiServer+"/pb.order/createRollback", createOrderReq).
		Add(stockRpcBusiServer+"/pb.stock/deduct", stockRpcBusiServer+"/pb.stock/deductRollback", deductReq)

	err = saga.Submit()
	fmt.Println("saga处理结果:", err)
	dtmimp.FatalIfError(err)
	if err != nil {
		return nil, fmt.Errorf("submit data to  dtm-server err  : %+v \n", err)
	}

	return &types.QuickCreateResp{}, nil
}
