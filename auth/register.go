package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
	"gosalusa.com/database/builder"
	"gosalusa.com/di"
	"gosalusa.com/internal/helpers"
)

type userRegisterDeps struct {
	DB     *sqlx.DB `inject:""`
	Claims *Claims  `inject:""`
}

func Register[T User](ctx context.Context) {
	di.Register(ctx, func(ctx context.Context, tag string) (*Claims, error) {
		c, _ := GetClaimsCtx(ctx)
		return c, nil
	})
	di.RegisterWith(ctx, func(ctx context.Context, tag string, deps *userRegisterDeps) (T, error) {
		var zero T
		if deps.Claims == nil {
			return zero, Err401Unauthorized
		}
		v, err := helpers.NewOf[T]()
		if err != nil {
			var zero T
			return zero, err
		}
		pkeyValues, err := helpers.PrimaryKeyValue(v)
		if err != nil {
			return zero, err
		}

		if len(pkeyValues) != 1 {
			return zero, fmt.Errorf("can only inject users with 1 primary key")
		}
		var pkey any
		switch pkeyValues[0].(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			err := json.Unmarshal([]byte(deps.Claims.Subject), &pkey)
			if err != nil {
				return zero, fmt.Errorf("invalid subject must be a number: %w", err)
			}
		default:
			pkey = deps.Claims.Subject
		}

		u, err := builder.From[T]().WithContext(ctx).Find(deps.DB, pkey)
		if err != nil {
			return zero, err
		}
		if reflect.ValueOf(u).IsZero() {
			return zero, Err401Unauthorized
		}
		return u, err
	})
}
