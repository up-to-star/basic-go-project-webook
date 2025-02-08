package integrationn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/basic-go-project-webook/webook/internal/integrationn/startup"
	"github.com/basic-go-project-webook/webook/internal/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUserHandler_SendSMSCode(t *testing.T) {
	server := startup.InitWebServer()
	rdb := startup.InitRedis()
	testCases := []struct {
		name     string
		before   func(*testing.T)
		after    func(*testing.T)
		phone    string
		wantCode int
		wantBody web.Result
	}{
		{
			name: "发送成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()
				key := "phone_code:login:15212345678"
				code, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, len(code) > 0)
				dur, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur > time.Minute*9+time.Second+50)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			phone:    "15212345678",
			wantCode: http.StatusOK,
			wantBody: web.Result{
				Msg: "发送成功",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			req, err := http.NewRequest(http.MethodPost,
				"/users/login_sms/code/send",
				bytes.NewReader([]byte(fmt.Sprintf(`{"phone": "%s"}`, tc.phone))))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			var result web.Result
			err = json.NewDecoder(recorder.Body).Decode(&result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantBody, result)
		})
	}
}
