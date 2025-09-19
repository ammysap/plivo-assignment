package usercontext

import (
	"context"
	"errors"

	"github.com/ammysap/plivo-pub-sub/logging"
	"github.com/gin-gonic/gin"
)

type SHKey string

var (
	UserKey    = SHKey("user")
	ServiceKey = SHKey("service")
)

type User struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type Service struct {
	ServiceName string `json:"service_name"`
}

func GetContextFromGinContext(c *gin.Context) (context.Context, error) {
	acAny, exists := c.Get("shcontext")

	if !exists {
		return nil, errors.New("sh context not found in gin context")
	}

	ac, ok := acAny.(context.Context)

	if !ok {
		// auth header not present
		return nil, errors.New("shcontext was not set by middleware")
	}

	return ac, nil
}

func CreateSHContextFromUserContext(
	ctx context.Context,
	sub *User,
	ser *Service,
) context.Context {
	shctx := context.WithValue(ctx, UserKey, sub)
	shctx = context.WithValue(shctx, ServiceKey, ser)
	shctx = logging.NewContext(
		shctx,
		sub.Phone,
		"",
		ser.ServiceName,
		sub.Email,
		false,
	)

	return shctx
}

func GetUserFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(UserKey).(*User)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return user, nil
}
