package randgen_test

import (
	"bytes"
	"fmt"
	"testing"

	"shortik/internal/core/service/randgen"
	"shortik/internal/core/service/randgen/model"
)

func TestGenerator_GenerateRandomBytes(t *testing.T) {
	type args struct {
		req model.GenerateRandomBytesRequest
	}
	tests := []struct {
		name          string
		args          args
		wantBufsCount int
		wantErr       error
	}{
		{
			name: "Normal",
			args: args{
				req: model.GenerateRandomBytesRequest{
					BufsCount: 3,
					Alphabet:  []byte("abc"),
					Len:       100,
				},
			},
			wantBufsCount: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := randgen.NewGenerator()
			got, err := g.GenerateRandomBytes(tt.args.req)
			if err := checkErrs(tt.wantErr, err); err != nil {
				t.Error(err)
				return
			}

			if len(got.Bufs) != tt.wantBufsCount {
				t.Errorf("expected to generate %d buffers, got %d", tt.wantBufsCount, len(got.Bufs))
				return
			}

			for i, buf := range got.Bufs {
				if len(buf) != tt.args.req.Len {
					t.Errorf("expected buffere %d length to be %d, got %d", i, tt.args.req.Len, len(buf))
					return
				}
				for i := range buf {
					if !bytes.Contains(tt.args.req.Alphabet, buf[i:i+1]) {
						t.Errorf("byte #%d (%d) is not in the generator's alphabet", i, buf[i])
						return
					}
				}
			}
		})
	}
}

func checkErrs(expectedErr error, actualErr error) error {
	if expectedErr == nil && actualErr == nil {
		return nil
	}
	if expectedErr == nil {
		return fmt.Errorf("expected nit error, got \"%w\"", actualErr)
	}
	if actualErr == nil {
		return fmt.Errorf("expected error \"%w\", got nil", expectedErr)
	}
	if expectedErr.Error() != actualErr.Error() {
		return fmt.Errorf("expected error: \"%w\", got: \"%w\"", expectedErr, actualErr)
	}
	return nil
}
