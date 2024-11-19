package dao

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name    string
		user    User
		mock    func(t *testing.T) *sql.DB
		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				res := sqlmock.NewResult(3, 1)
				mock.ExpectExec("INSERT INTO `users` .*").WillReturnResult(res)
				return mockDB
			},
			user: User{
				Email: sql.NullString{
					String: "test@example.com",
					Valid:  true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      tc.mock(t),
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			d := NewUserDAO(db)
			err = d.Insert(context.Background(), tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
