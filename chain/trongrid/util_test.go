package trongrid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		addr    string
		wantErr string
	}{
		{
			name:    "valid mainnet address",
			addr:    "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
			wantErr: "",
		},
		{
			name:    "valid testnet address",
			addr:    "TG3XXyExBkPp9nzdajDZsozEu4BkaSJozs",
			wantErr: "",
		},
		{
			name:    "empty address",
			addr:    "",
			wantErr: "invalid address length: got 0, want 34",
		},
		{
			name:    "too short",
			addr:    "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj",
			wantErr: "invalid address length: got 32, want 34",
		},
		{
			name:    "too long",
			addr:    "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6tAB",
			wantErr: "invalid address length: got 36, want 34",
		},
		{
			name:    "invalid prefix",
			addr:    "AR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
			wantErr: "invalid address prefix: got 'A', want 'T'",
		},
		{
			name:    "invalid checksum - modified last char",
			addr:    "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6u",
			wantErr: "invalid address checksum",
		},
		{
			name:    "invalid checksum - modified middle char",
			addr:    "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLX6t",
			wantErr: "invalid address checksum",
		},
		{
			name:    "invalid base58 - contains zero",
			addr:    "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgj0j6t",
			wantErr: "invalid decoded address length: got 0, want 25",
		},
		{
			name:    "invalid base58 - contains capital O",
			addr:    "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjOj6t",
			wantErr: "invalid decoded address length: got 0, want 25",
		},
		{
			name:    "invalid base58 - contains capital I",
			addr:    "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjIj6t",
			wantErr: "invalid decoded address length: got 0, want 25",
		},
		{
			name:    "invalid base58 - contains lowercase l",
			addr:    "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjlj6t",
			wantErr: "invalid decoded address length: got 0, want 25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateAddr(tt.addr)

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
			}
		})
	}
}
