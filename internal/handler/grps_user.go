package handler

import (
	"context"
	"crypto/rsa"

	"github.com/ObjoradDdd/AuthService/internal/model"
	"github.com/ObjoradDdd/AuthService/internal/service"
	pb "github.com/ObjoradDdd/AuthService/proto"
)

type AuthGRPCServer struct {
	pb.UnimplementedAuthServiceServer
	userService *service.UserService
	privateKey  *rsa.PrivateKey
}

func NewAuthGRPCServer(us *service.UserService, pk *rsa.PrivateKey) *AuthGRPCServer {
	return &AuthGRPCServer{userService: us, privateKey: pk}
}

func (s *AuthGRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.AuthResponse, error) {

	user, err := s.userService.RegisterUser(ctx, &model.User{Login: req.Login}, req.Password)
	if err != nil {
		return nil, err
	}

	token, err := generateToken(user.Id, s.privateKey)
	if err != nil {
		return nil, err
	}

	return &pb.AuthResponse{Token: token, UserId: int32(user.Id)}, nil
}

func (s *AuthGRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {

	user, token, err := s.userService.Login(ctx, &model.User{Login: req.Login}, req.Password)
	if err != nil {
		return nil, err
	}

	return &pb.AuthResponse{Token: token, UserId: int32(user.Id)}, nil
}

func (s *AuthGRPCServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.SuccesMessage, error) {

	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.userService.DeleteUserByID(ctx, userID); err != nil {
		return nil, err
	}

	return &pb.SuccesMessage{M: "user deleted"}, nil
}

func (s *AuthGRPCServer) ChangeUserPassword(ctx context.Context, req *pb.ChangeUserPasswordRequest) (*pb.SuccesMessage, error) {

	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.userService.UpdateUserHash(ctx, &model.User{Id: userID}, req.NewPassword, req.CurrentPassword); err != nil {
		return nil, err
	}

	return &pb.SuccesMessage{M: "password changed"}, nil
}

func (s *AuthGRPCServer) ChangeUserLogin(ctx context.Context, req *pb.ChangeUserLoginRequest) (*pb.SuccesMessage, error) {

	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.userService.UpdateUserLogin(ctx, &model.User{Id: userID, Login: req.NewLogin}); err != nil {
		return nil, err
	}

	return &pb.SuccesMessage{M: "login changed"}, nil
}
