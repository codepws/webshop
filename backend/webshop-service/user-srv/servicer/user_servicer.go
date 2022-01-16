// 包注释
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
// @Update  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
package servicer

import (
	"context"
	"errors"
	"log"
	"webshop-service/user-srv/dao"

	pb "webshop-service/user-srv/proto"
)

// greeterServer 定义一个结构体用于实现 .proto文件中定义的方法
// 新版本 gRPC 要求必须嵌入 pb.UnimplementedGreeterServer 结构体
type UserServicer struct {
	pb.UnimplementedUserServer
}

// 注册用户
func (user *UserServicer) RegisterUser(ctx context.Context, request *pb.UserRegisterRequest) (*pb.UserBaseInfoResponse, error) {

	log.Println("服务方法[CreateUser]：新建用户")

	//新建用户, 表单验证，没有必要

	userID, err := dao.RegisterUser(request)
	if err != nil {
		return nil, err
	}

	rsp := &pb.UserBaseInfoResponse{
		Id:       userID,
		Nickname: request.Nickname,
		Role:     0,
		//HeadUrl:  "",
	}

	return rsp, nil
}

// 用户登录
func (user *UserServicer) LoginUser(ctx context.Context, request *pb.UserLoginRequest) (*pb.UserBaseInfoResponse, error) {

	log.Println("服务方法[LoginUser]用户登录")

	//新建用户, 表单验证，没有必要

	userinfo, err := dao.LoginUser(request)
	if err != nil {
		return nil, errors.New(err.ToString())
	}

	return userinfo, nil
}

func (user *UserServicer) GetUserList(ctx context.Context, request *pb.PageInfo) (*pb.UserListResonse, error) {

	log.Println("服务方法[GetUserList]：获取用户列表")

	/*
		users = User.select()
		rsp.total = users.count()
		print("用户列表")
		start = 0
		per_page_nums = 10
		if request.pSize:
			per_page_nums = request.pSize
		if request.pn:
			start = per_page_nums * (request.pn - 1)

		users = users.limit(per_page_nums).offset(start)

		for user in users:
			rsp.data.append(self.convert_user_to_rsp(user))
	*/
	//Total int32               `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	//Data  []*UserInfoResponse
	// 获取用户的列表
	rsp := &pb.UserListResonse{Total: 123, Data: nil}
	return rsp, nil
}

// SayHello 简单实现一下.proto文件中定义的 SayHello 方法
func (user *UserServicer) UnaryEcho(ctx context.Context, request *pb.EchoRequest) (*pb.EchoResponse, error) {
	log.Printf("普通一元方法，客户端请求：%v", request.GetName())

	//返回消息
	return &pb.EchoResponse{
		Message: "Hello, " + request.Name,
	}, nil

	//返回错误
	//return nil, status.Errorf(codes.NotFound, "记录未找到：%s", request.Name)
}
