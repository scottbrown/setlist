package setlist

import (
	"errors"
	"testing"
)

func TestNewIdentityStoreId(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "valid identity store id",
			input:   "d-1234567890",
			want:    "d-1234567890",
			wantErr: nil,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: ErrEmptyString,
		},
		{
			name:    "missing d- prefix",
			input:   "1234567890",
			want:    "",
			wantErr: ErrWrongFormat,
		},
		{
			name:    "wrong prefix",
			input:   "x-1234567890",
			want:    "",
			wantErr: ErrWrongFormat,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewIdentityStoreId(tc.input)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("NewIdentityStoreId(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}

			if got.String() != tc.want {
				t.Errorf("NewIdentityStoreId(%q) = %q, want %q", tc.input, got.String(), tc.want)
			}
		})
	}
}

func TestNewRegion(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "valid region",
			input:   "ca-central-1",
			want:    "ca-central-1",
			wantErr: nil,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: ErrEmptyString,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewRegion(tc.input)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("NewRegion(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}

			if got.String() != tc.want {
				t.Errorf("NewRegion(%q) = %q, want %q", tc.input, got.String(), tc.want)
			}
		})
	}
}
