package service

import (
	"context"
	"time"

	pb "github.com/storm/myidea/api/user/v1"
	"github.com/storm/myidea/service/user/internal/biz"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
	userBiz    *biz.UserBiz
	addressBiz *biz.AddressBiz
}

func NewUserService(userBiz *biz.UserBiz, addressBiz *biz.AddressBiz) *UserService {
	return &UserService{
		userBiz:    userBiz,
		addressBiz: addressBiz,
	}
}

func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := s.userBiz.Register(ctx, req.Username, req.Password, req.Phone, req.Email, req.Role)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{User: userToProto(user)}, nil
}

func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.userBiz.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{User: userToProto(user)}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserProto, error) {
	user, err := s.userBiz.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return userToProto(user), nil
}

func (s *UserService) CreateAddress(ctx context.Context, req *pb.CreateAddressRequest) (*pb.AddressProto, error) {
	addr := &biz.Address{
		UserID:        req.UserId,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		IsDefault:     req.IsDefault,
	}

	id, err := s.addressBiz.Create(ctx, addr)
	if err != nil {
		return nil, err
	}
	addr.ID = id
	return addressToProto(addr), nil
}

func (s *UserService) UpdateAddress(ctx context.Context, req *pb.UpdateAddressRequest) (*pb.AddressProto, error) {
	addr := &biz.Address{
		ID:            req.Id,
		UserID:        req.UserId,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		IsDefault:     req.IsDefault,
	}

	if err := s.addressBiz.Update(ctx, addr); err != nil {
		return nil, err
	}
	return addressToProto(addr), nil
}

func (s *UserService) DeleteAddress(ctx context.Context, req *pb.DeleteAddressRequest) (*pb.Empty, error) {
	if err := s.addressBiz.Delete(ctx, req.Id, req.UserId); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *UserService) ListAddresses(ctx context.Context, req *pb.ListAddressesRequest) (*pb.ListAddressesResponse, error) {
	addrs, err := s.addressBiz.ListByUserID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	protoAddrs := make([]*pb.AddressProto, len(addrs))
	for i := range addrs {
		protoAddrs[i] = addressToProto(addrs[i])
	}
	return &pb.ListAddressesResponse{Addresses: protoAddrs}, nil
}

func (s *UserService) SetDefaultAddress(ctx context.Context, req *pb.SetDefaultAddressRequest) (*pb.AddressProto, error) {
	if err := s.addressBiz.SetDefault(ctx, req.Id, req.UserId); err != nil {
		return nil, err
	}
	return &pb.AddressProto{Id: req.Id}, nil
}

func (s *UserService) AdminListUsers(ctx context.Context, req *pb.AdminListUsersRequest) (*pb.AdminListUsersResponse, error) {
	users, total, err := s.userBiz.AdminListUsers(ctx, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	protos := make([]*pb.UserProto, len(users))
	for i, u := range users {
		protos[i] = userToProto(u)
	}
	return &pb.AdminListUsersResponse{Users: protos, Total: total}, nil
}

func userToProto(u *biz.User) *pb.UserProto {
	return &pb.UserProto{
		Id:        u.ID,
		Username:  u.Username,
		Phone:     u.Phone,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}

func addressToProto(a *biz.Address) *pb.AddressProto {
	return &pb.AddressProto{
		Id:            a.ID,
		UserId:        a.UserID,
		ReceiverName:  a.ReceiverName,
		ReceiverPhone: a.ReceiverPhone,
		Province:      a.Province,
		City:          a.City,
		District:      a.District,
		DetailAddress: a.DetailAddress,
		IsDefault:     a.IsDefault,
		CreatedAt:     a.CreatedAt.Format(time.RFC3339),
	}
}

var _ pb.UserServiceServer = (*UserService)(nil)
