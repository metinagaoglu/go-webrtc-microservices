package routes

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/metinagaoglu/go-grpc-api-gateway/pkg/auth/pb"
)

type RegisterRequestBody struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(ctx *gin.Context, c pb.AuthServiceClient) {
	body := RegisterRequestBody{}

	if err := ctx.BindJSON(&body); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	res, err := c.Register(context.Background(), &pb.RegisterRequest{
		Email:    body.Email,
		Password: body.Password,
	})

	if err != nil {
		ctx.AbortWithError(http.StatusBadGateway, err)
		return
	}

	ctx.JSON(int(res.Status), &res)
}
