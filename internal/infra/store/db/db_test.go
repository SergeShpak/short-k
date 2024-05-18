package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"shortik/internal/infra/store/db/internal/mocks"
	"shortik/internal/infra/store/db/internal/queries"
	"shortik/internal/infra/store/db/model"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/mock/gomock"
)

func TestDB_StoreURL(t *testing.T) {
	tests := []struct {
		name             string
		req              model.StoreURLRequest
		handlerResp      queries.InsertURLRow
		handlerErr       error
		want             model.StoreURLResponse
		expectedErr      error
		expectedErrCheck areErrsEqualFn
	}{
		{
			name: "normal",
			req: model.StoreURLRequest{
				URL:  "example.com",
				Slug: "42",
			},
			handlerResp: queries.InsertURLRow{
				Url:  "example.com",
				Slug: "42",
			},
			handlerErr: nil,
			want: model.StoreURLResponse{
				URL:               "example.com",
				Slug:              "42",
				IsNewSlugInserted: true,
			},
		},
		{
			name: "URL already exists",
			req: model.StoreURLRequest{
				URL:  "example.com",
				Slug: "42",
			},
			handlerResp: queries.InsertURLRow{
				Url:  "example.com",
				Slug: "24",
			},
			handlerErr: nil,
			want: model.StoreURLResponse{
				URL:               "example.com",
				Slug:              "24",
				IsNewSlugInserted: false,
			},
		},
		{
			name: "slug already exists",
			req: model.StoreURLRequest{
				URL:  "example.com",
				Slug: "42",
			},
			handlerResp: queries.InsertURLRow{},
			handlerErr: &pgconn.PgError{
				Code:           pgerrcode.UniqueViolation,
				ConstraintName: "unique_slug",
			},
			want:             model.StoreURLResponse{},
			expectedErr:      model.ErrSlugAlreadyExists,
			expectedErrCheck: areEqualTypedErrors,
		},
		{
			name: "another unique violation error",
			req: model.StoreURLRequest{
				URL:  "example.com",
				Slug: "42",
			},
			handlerResp: queries.InsertURLRow{},
			handlerErr: &pgconn.PgError{
				Code:           pgerrcode.UniqueViolation,
				ConstraintName: "non_existent_constraint",
			},
			want:             model.StoreURLResponse{},
			expectedErr:      errors.New("failed to store the URL: :  (SQLSTATE 23505)"),
			expectedErrCheck: areEqualGenericErrors,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := mocks.NewMockhandler(ctrl)
			h.EXPECT().
				InsertURL(gomock.Any(), queries.InsertURLParams{
					Url:  string(tt.req.URL),
					Slug: string(tt.req.Slug),
				}).
				Times(1).
				Return(tt.handlerResp, tt.handlerErr)

			db := &DB{
				handler: h,
			}

			got, err := db.StoreURL(context.Background(), tt.req)
			if err := checkErrs(tt.expectedErr, err, tt.expectedErrCheck); err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB.StoreURL() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func TestDB_GetURL(t *testing.T) {
	tests := []struct {
		name             string
		req              model.GetURLRequest
		handlerResp      string
		handlerErr       error
		want             model.GetURLResponse
		expectedErr      error
		expectedErrCheck areErrsEqualFn
	}{
		{
			name: "normal",
			req: model.GetURLRequest{
				Slug: "42",
			},
			handlerResp: "example.com",
			handlerErr:  nil,
			want: model.GetURLResponse{
				FullURL: "example.com",
			},
		},
		{
			name: "no rows",
			req: model.GetURLRequest{
				Slug: "42",
			},
			handlerResp:      "example.com",
			handlerErr:       pgx.ErrNoRows,
			want:             model.GetURLResponse{},
			expectedErr:      model.ErrSlugNotFound,
			expectedErrCheck: areEqualTypedErrors,
		},
		{
			name: "generic error",
			req: model.GetURLRequest{
				Slug: "42",
			},
			handlerResp:      "example.com",
			handlerErr:       errors.New("something went wrong"),
			want:             model.GetURLResponse{},
			expectedErr:      errors.New("failed to get a URL by slug 42: something went wrong"),
			expectedErrCheck: areEqualGenericErrors,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			h := mocks.NewMockhandler(ctrl)
			h.EXPECT().
				GetURL(gomock.Any(), string(tt.req.Slug)).
				Times(1).
				Return(tt.handlerResp, tt.handlerErr)

			db := &DB{
				handler: h,
			}

			got, err := db.GetURL(context.Background(), tt.req)
			if err := checkErrs(tt.expectedErr, err, tt.expectedErrCheck); err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB.StoreURL() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

type areErrsEqualFn func(expectedErr error, actualErr error) error

func checkErrs(expectedErr error, actualErr error, areEqual areErrsEqualFn) error {
	if expectedErr == nil && actualErr == nil {
		return nil
	}
	if expectedErr == nil {
		return fmt.Errorf("expected nit error, got \"%w\"", actualErr)
	}
	if actualErr == nil {
		return fmt.Errorf("expected error \"%w\", got nil", expectedErr)
	}
	if areEqual == nil {
		areEqual = areEqualGenericErrors
	}
	if err := areEqual(expectedErr, actualErr); err != nil {
		return err
	}
	return nil
}

func areEqualGenericErrors(expectedErr error, actualErr error) error {
	if expectedErr.Error() != actualErr.Error() {
		return fmt.Errorf("expected error: \"%w\", got: \"%w\"", expectedErr, actualErr)
	}
	return nil
}

func areEqualTypedErrors(expectedErr error, actualErr error) error {
	if !errors.Is(actualErr, expectedErr) {
		return fmt.Errorf("expected error \"%w\" and actual error \"%w\" have different types", expectedErr, actualErr)
	}
	return nil
}
