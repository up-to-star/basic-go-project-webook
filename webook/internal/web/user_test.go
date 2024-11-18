package web

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserHandle_Signup(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{},
	}

	req, err := http.NewRequest(http.MethodPost, "/user/signup", bytes.NewBuffer([]byte(`
{
	"email": "test@example.com",
	"password": "password"
}`)))
	t.Log(req)
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	t.Log(resp)
	//h := NewUserHandle()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
