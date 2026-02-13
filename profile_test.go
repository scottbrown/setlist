package setlist

import (
	"errors"
	"testing"
)

func TestNewProfileName(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "valid profile name",
			input:   "123456789012-AdminAccess",
			want:    "123456789012-AdminAccess",
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
			got, err := NewProfileName(tc.input)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("NewProfileName(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}

			if got.String() != tc.want {
				t.Errorf("NewProfileName(%q) = %q, want %q", tc.input, got.String(), tc.want)
			}
		})
	}
}

func TestNewProfileDescription(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "valid description",
			input:   "Admin access for production",
			want:    "Admin access for production",
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
			got, err := NewProfileDescription(tc.input)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("NewProfileDescription(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}

			if got.String() != tc.want {
				t.Errorf("NewProfileDescription(%q) = %q, want %q", tc.input, got.String(), tc.want)
			}
		})
	}
}

func TestNewSessionDuration(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "valid duration",
			input:   "PT1H",
			want:    "PT1H",
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
			got, err := NewSessionDuration(tc.input)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("NewSessionDuration(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}

			if got.String() != tc.want {
				t.Errorf("NewSessionDuration(%q) = %q, want %q", tc.input, got.String(), tc.want)
			}
		})
	}
}

func TestNewSessionName(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "valid session name",
			input:   "my-sso-session",
			want:    "my-sso-session",
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
			got, err := NewSessionName(tc.input)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("NewSessionName(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}

			if got.String() != tc.want {
				t.Errorf("NewSessionName(%q) = %q, want %q", tc.input, got.String(), tc.want)
			}
		})
	}
}

func TestNewAWSAccountId(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "valid 12-digit account id",
			input:   "123456789012",
			want:    "123456789012",
			wantErr: nil,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: ErrEmptyString,
		},
		{
			name:    "too short",
			input:   "12345",
			want:    "",
			wantErr: ErrInvalidAWSAccountIdLength,
		},
		{
			name:    "too long",
			input:   "1234567890123",
			want:    "",
			wantErr: ErrInvalidAWSAccountIdLength,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewAWSAccountId(tc.input)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("NewAWSAccountId(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}

			if got.String() != tc.want {
				t.Errorf("NewAWSAccountId(%q) = %q, want %q", tc.input, got.String(), tc.want)
			}
		})
	}
}

func TestNewRoleName(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "valid role name",
			input:   "AdministratorAccess",
			want:    "AdministratorAccess",
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
			got, err := NewRoleName(tc.input)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("NewRoleName(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}

			if got.String() != tc.want {
				t.Errorf("NewRoleName(%q) = %q, want %q", tc.input, got.String(), tc.want)
			}
		})
	}
}
