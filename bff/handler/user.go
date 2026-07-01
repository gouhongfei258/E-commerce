package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/storm/myidea/api/user/v1"
)

// UserHandler handles HTTP requests for user auth and address management.
type UserHandler struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
	jwtSecret string
	jwtExpiry time.Duration
}

// NewUserHandler creates a new handler and dials the gRPC connection.
func NewUserHandler(grpcAddr string, dialTimeout time.Duration, jwtSecret string, jwtExpiry time.Duration) (*UserHandler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryClientInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	return &UserHandler{
		client:    pb.NewUserServiceClient(conn),
		conn:      conn,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}, nil
}

// Close shuts down the gRPC connection.
func (h *UserHandler) Close() error {
	return h.conn.Close()
}

func (h *UserHandler) UserServiceClient() pb.UserServiceClient {
	return h.client
}

// Register  POST /api/v1/auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=64"`
		Password string `json:"password" binding:"required,min=6"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.client.Register(ctx, &pb.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Phone:    req.Phone,
		Email:    req.Email,
		Role:     "user",
	})
	if err != nil {
		respondError(c, err)
		return
	}

	token, err := h.generateJWT(resp.User.Id, resp.User.Role)
	if err != nil {
		respond(c, http.StatusInternalServerError, 500, "failed to generate token", nil)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{
		"user":  resp.User,
		"token": token,
	})
}

// Login  POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := c.Request.Context()
	resp, err := h.client.Login(ctx, &pb.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	token, err := h.generateJWT(resp.User.Id, resp.User.Role)
	if err != nil {
		respond(c, http.StatusInternalServerError, 500, "failed to generate token", nil)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{
		"user":  resp.User,
		"token": token,
	})
}

// ListAddresses  GET /api/v1/addresses
func (h *UserHandler) ListAddresses(c *gin.Context) {
	userID := c.GetInt64("user_id")
	ctx := injectUserID(c, userID)

	resp, err := h.client.ListAddresses(ctx, &pb.ListAddressesRequest{UserId: userID})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", gin.H{"addresses": resp.Addresses})
}

// CreateAddress  POST /api/v1/addresses
func (h *UserHandler) CreateAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ReceiverName  string `json:"receiver_name" binding:"required"`
		ReceiverPhone string `json:"receiver_phone" binding:"required"`
		Province      string `json:"province"`
		City          string `json:"city"`
		District      string `json:"district"`
		DetailAddress string `json:"detail_address" binding:"required"`
		IsDefault     bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := injectUserID(c, userID)
	resp, err := h.client.CreateAddress(ctx, &pb.CreateAddressRequest{
		UserId:        userID,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		IsDefault:     req.IsDefault,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp)
}

// UpdateAddress  PUT /api/v1/addresses/:id
func (h *UserHandler) UpdateAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid address id", nil)
		return
	}

	var req struct {
		ReceiverName  string `json:"receiver_name" binding:"required"`
		ReceiverPhone string `json:"receiver_phone" binding:"required"`
		Province      string `json:"province"`
		City          string `json:"city"`
		District      string `json:"district"`
		DetailAddress string `json:"detail_address" binding:"required"`
		IsDefault     bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respond(c, http.StatusBadRequest, 400, err.Error(), nil)
		return
	}

	ctx := injectUserID(c, userID)
	resp, err := h.client.UpdateAddress(ctx, &pb.UpdateAddressRequest{
		Id:            id,
		UserId:        userID,
		ReceiverName:  req.ReceiverName,
		ReceiverPhone: req.ReceiverPhone,
		Province:      req.Province,
		City:          req.City,
		District:      req.District,
		DetailAddress: req.DetailAddress,
		IsDefault:     req.IsDefault,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", resp)
}

// DeleteAddress  DELETE /api/v1/addresses/:id
func (h *UserHandler) DeleteAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid address id", nil)
		return
	}

	ctx := injectUserID(c, userID)
	_, err = h.client.DeleteAddress(ctx, &pb.DeleteAddressRequest{
		Id:     id,
		UserId: userID,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", nil)
}

// SetDefaultAddress  PUT /api/v1/addresses/:id/default
func (h *UserHandler) SetDefaultAddress(c *gin.Context) {
	userID := c.GetInt64("user_id")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respond(c, http.StatusBadRequest, 400, "invalid address id", nil)
		return
	}

	ctx := injectUserID(c, userID)
	_, err = h.client.SetDefaultAddress(ctx, &pb.SetDefaultAddressRequest{
		Id:     id,
		UserId: userID,
	})
	if err != nil {
		respondError(c, err)
		return
	}

	respond(c, http.StatusOK, 0, "ok", nil)
}

// generateJWT creates a JWT token with user_id and role claims.
func (h *UserHandler) generateJWT(userID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(h.jwtExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
