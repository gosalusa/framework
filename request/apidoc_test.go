package request

import (
	"context"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/model"
	"gosalusa.com/openapidoc"
)

type apiTestModel struct {
	model.BaseModel
	ID int `db:"id,primary"`
}

type apiTestRequest struct {
	Name string        `json:"name"`
	Q    int           `query:"q"`
	ID   string        `path:"id"`
	Foo  *apiTestModel `inject:"foo"`
}

type apiTestResponse struct {
	Value string `json:"value"`
}

type apiBadRequest struct {
	C chan int `json:"c"`
}

type apiCTResponse struct {
	Value string `json:"value"`
}

type apiBadQueryRequest struct {
	C chan int `query:"c"`
}

type apiBadPathRequest struct {
	C chan int `path:"c"`
}

type apiBadModel struct {
	model.BaseModel
	ID chan int `db:"id,primary"`
}

type apiBadModelRequest struct {
	M *apiBadModel `inject:"m"`
}

func TestOperation(t *testing.T) {
	t.Run("no docs", func(t *testing.T) {
		h := Handler(func(r *apiTestRequest) (*apiTestResponse, error) {
			return nil, nil
		})
		op, err := h.Operation(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, op.Description)
		assert.NotNil(t, op.Responses)
		assert.NotNil(t, op.Responses.Default)
		assert.Contains(t, op.Produces, "application/json")
		assert.Len(t, op.Parameters, 4)
	})

	t.Run("with docs", func(t *testing.T) {
		h := Handler(func(r *apiTestRequest) (*apiTestResponse, error) {
			return nil, nil
		}).Docs(&spec.OperationProps{
			Description: "custom description",
		})
		op, err := h.Operation(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "custom description", op.Description)
	})

	t.Run("schema error", func(t *testing.T) {
		h := Handler(func(r *apiBadRequest) (*apiTestResponse, error) {
			return nil, nil
		})
		_, err := h.Operation(context.Background())
		assert.Error(t, err)
	})

	t.Run("registered content type", func(t *testing.T) {
		openapidoc.RegisterContentType[apiCTResponse]("application/vnd.api+json")
		h := Handler(func(r *apiTestRequest) (apiCTResponse, error) {
			return apiCTResponse{}, nil
		})
		op, err := h.Operation(context.Background())
		assert.NoError(t, err)
		assert.Contains(t, op.Produces, "application/vnd.api+json")
	})
}

func TestNewAPIRequest(t *testing.T) {
	params, err := newAPIRequest[apiTestRequest]()
	assert.NoError(t, err)
	assert.Len(t, params, 4)

	names := map[string]string{}
	for i := range params {
		names[params[i].Name] = params[i].In
	}
	assert.Equal(t, "query", names["q"])
	assert.Equal(t, "path", names["id"])
	assert.Equal(t, "path", names["foo"])
	assert.Equal(t, "apiTestRequest", params[0].Name)

	_, err = newAPIRequest[apiBadRequest]()
	assert.Error(t, err)
}

func TestNewAPIRequestErrors(t *testing.T) {
	_, err := newAPIRequest[apiBadQueryRequest]()
	assert.Error(t, err)

	_, err = newAPIRequest[apiBadPathRequest]()
	assert.Error(t, err)

	_, err = newAPIRequest[apiBadModelRequest]()
	assert.Error(t, err)
}
