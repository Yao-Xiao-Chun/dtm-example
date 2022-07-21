package logic

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"gozerodtm/order-srv/internal/model"
	"gozerodtm/order-srv/internal/svc"
	"gozerodtm/order-srv/pb"

	"github.com/yedf/dtmgrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tal-tech/go-zero/core/logx"
)

type CreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateLogic {
	return &CreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateLogic) Create(in *pb.CreateReq) (*pb.CreateResp, error) {

	fmt.Printf("创建订单 in : %+v \n", in)

	//barrier防止空补偿、空悬挂等具体看dtm官网即可，别忘记加barrier表在当前库中，因为判断补偿与要执行的sql一起本地事务
	//空补偿 cancel执行时，try未执行，事务分支的cancel操作需要判断出try未执行，这时需要忽略cancel中的业务数据更新，直接返回
	//空悬挂 try执行，cancel已执行完成，事务分支的try操作需要判断出cancel已执行，这时需要忽略try中的业务数据更新，直接返回
	//幂等  由于任何一个请求都可能出现网络异常，出现重复请求，所有的分布式事务分支操作，都需要保证幂等性
	barrier, err := dtmgrpc.BarrierFromGrpc(l.ctx)

	db, err := sqlx.NewMysql(l.svcCtx.Config.DB.DataSource).RawDB() //创建一个本地db事务

	if err != nil {
		//!!!一般数据库不会错误不需要dtm回滚，就让他一直重试，这时候就不要返回codes.Aborted, dtmcli.ResultFailure 就可以了，具体自己把控!!!
		return nil, status.Error(codes.Internal, err.Error()) //返回这个，dtm才会进行回滚
	}
	//子事务屏障 解决空补偿 空悬挂 幂等
	if err := barrier.CallWithDB(db, func(tx *sql.Tx) error {

		order := new(model.Order)
		order.GoodsId = in.GoodsId
		order.Num = in.Num
		order.UserId = in.UserId

		_, err = l.svcCtx.OrderModel.Insert(tx, order)
		if err != nil {
			return fmt.Errorf("创建订单失败 err : %v , order:%+v \n", err, order)
		}

		return nil
	}); err != nil {
		//!!!一般数据库不会错误不需要dtm回滚，就让他一直重试，这时候就不要返回codes.Aborted, dtmcli.ResultFailure 就可以了，具体自己把控!!!
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateResp{}, nil
}
