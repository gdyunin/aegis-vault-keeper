package bankcard

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBankCard(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	validCardNumber := "4532015112830366" // Valid test card number that passes Luhn

	type args struct {
		params *NewBankCardParams
	}
	tests := []struct {
		args    args
		want    func(t *testing.T, bc *BankCard)
		name    string
		wantErr bool
	}{
		{
			name: "valid/complete_card",
			args: args{
				params: &NewBankCardParams{
					CardNumber:  validCardNumber,
					CardHolder:  "John Doe",
					ExpiryMonth: "12",
					ExpiryYear:  "2025",
					CVV:         "123",
					Description: "Personal credit card",
					UserID:      userID,
				},
			},
			want: func(t *testing.T, bc *BankCard) {
				t.Helper()
				assert.NotEqual(t, uuid.Nil, bc.ID)
				assert.Equal(t, userID, bc.UserID)
				assert.Equal(t, []byte(validCardNumber), bc.CardNumber)
				assert.Equal(t, []byte("John Doe"), bc.CardHolder)
				assert.Equal(t, []byte("12"), bc.ExpiryMonth)
				assert.Equal(t, []byte("2025"), bc.ExpiryYear)
				assert.Equal(t, []byte("123"), bc.CVV)
				assert.Equal(t, []byte("Personal credit card"), bc.Description)
				assert.WithinDuration(t, time.Now(), bc.UpdatedAt, time.Second)
			},
		},
		{
			name: "valid/minimal_card",
			args: args{
				params: &NewBankCardParams{
					CardNumber:  validCardNumber,
					CardHolder:  "Jane Smith",
					ExpiryMonth: "01",
					ExpiryYear:  "2026",
					CVV:         "456",
					Description: "",
					UserID:      userID,
				},
			},
			want: func(t *testing.T, bc *BankCard) {
				t.Helper()
				assert.NotEqual(t, uuid.Nil, bc.ID)
				assert.Equal(t, userID, bc.UserID)
				assert.Equal(t, []byte(validCardNumber), bc.CardNumber)
				assert.Equal(t, []byte("Jane Smith"), bc.CardHolder)
				assert.Equal(t, []byte("01"), bc.ExpiryMonth)
				assert.Equal(t, []byte("2026"), bc.ExpiryYear)
				assert.Equal(t, []byte("456"), bc.CVV)
				assert.Equal(t, []byte(""), bc.Description)
			},
		},
		{
			name: "invalid/card_validation_fails",
			args: args{
				params: &NewBankCardParams{
					CardNumber:  "invalid",
					CardHolder:  "",
					ExpiryMonth: "13",
					ExpiryYear:  "2020",
					CVV:         "12",
					UserID:      userID,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid/nil_params",
			args: args{
				params: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.args.params == nil {
				// Test nil params separately since it would panic
				assert.Panics(t, func() {
					_, _ = NewBankCard(tt.args.params)
				})
				return
			}

			got, err := NewBankCard(tt.args.params)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			if tt.want != nil {
				tt.want(t, got)
			}
		})
	}
}

func TestNewBankCardParams_Validate(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	validCardNumber := "4532015112830366"

	tests := []struct {
		errType error
		params  *NewBankCardParams
		name    string
		wantErr bool
	}{
		{
			name: "valid/all_fields_correct",
			params: &NewBankCardParams{
				CardNumber:  validCardNumber,
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "123",
				Description: "Test card",
				UserID:      userID,
			},
			wantErr: false,
		},
		{
			name: "valid/empty_description",
			params: &NewBankCardParams{
				CardNumber:  validCardNumber,
				CardHolder:  "Jane Smith",
				ExpiryMonth: "01",
				ExpiryYear:  "2026",
				CVV:         "456",
				Description: "",
				UserID:      userID,
			},
			wantErr: false,
		},
		{
			name: "invalid/card_number_non_digits",
			params: &NewBankCardParams{
				CardNumber:  "4532-0151-1283-0366",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "123",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrInvalidCardNumber,
		},
		{
			name: "invalid/card_number_too_short",
			params: &NewBankCardParams{
				CardNumber:  "123456789012",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "123",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrInvalidCardNumber,
		},
		{
			name: "invalid/card_number_too_long",
			params: &NewBankCardParams{
				CardNumber:  "12345678901234567890",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "123",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrInvalidCardNumber,
		},
		{
			name: "invalid/card_number_luhn_check_fails",
			params: &NewBankCardParams{
				CardNumber:  "4532015112830367", // Last digit changed to fail Luhn
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "123",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrLuhnCheckFailed,
		},
		{
			name: "invalid/empty_card_holder",
			params: &NewBankCardParams{
				CardNumber:  validCardNumber,
				CardHolder:  "",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "123",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrEmptyCardHolder,
		},
		{
			name: "invalid/expiry_month_invalid_format",
			params: &NewBankCardParams{
				CardNumber:  validCardNumber,
				CardHolder:  "John Doe",
				ExpiryMonth: "1",
				ExpiryYear:  "2025",
				CVV:         "123",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrInvalidExpiryMonth,
		},
		{
			name: "invalid/expiry_month_out_of_range",
			params: &NewBankCardParams{
				CardNumber:  validCardNumber,
				CardHolder:  "John Doe",
				ExpiryMonth: "13",
				ExpiryYear:  "2025",
				CVV:         "123",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrInvalidExpiryMonth,
		},
		{
			name: "invalid/expiry_year_invalid_format",
			params: &NewBankCardParams{
				CardNumber:  validCardNumber,
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "25",
				CVV:         "123",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrInvalidExpiryYear,
		},
		{
			name: "invalid/card_expired_year",
			params: &NewBankCardParams{
				CardNumber:  validCardNumber,
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2020",
				CVV:         "123",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrCardExpired,
		},
		{
			name: "invalid/cvv_too_short",
			params: &NewBankCardParams{
				CardNumber:  validCardNumber,
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "12",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrInvalidCVV,
		},
		{
			name: "invalid/cvv_too_long",
			params: &NewBankCardParams{
				CardNumber:  validCardNumber,
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "12345",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrInvalidCVV,
		},
		{
			name: "invalid/cvv_non_digits",
			params: &NewBankCardParams{
				CardNumber:  validCardNumber,
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "abc",
				UserID:      userID,
			},
			wantErr: true,
			errType: ErrInvalidCVV,
		},
		{
			name: "invalid/multiple_errors",
			params: &NewBankCardParams{
				CardNumber:  "invalid",
				CardHolder:  "",
				ExpiryMonth: "13",
				ExpiryYear:  "20",
				CVV:         "1",
				UserID:      userID,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLuhnValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "valid/visa_test_card",
			number: "4532015112830366",
			want:   true,
		},
		{
			name:   "valid/mastercard_test_card",
			number: "5555555555554444",
			want:   true,
		},
		{
			name:   "valid/amex_test_card",
			number: "371449635398431",
			want:   true,
		},
		{
			name:   "valid/discover_test_card",
			number: "6011111111111117",
			want:   true,
		},
		{
			name:   "invalid/wrong_checksum",
			number: "4532015112830367",
			want:   false,
		},
		{
			name:   "invalid/all_zeros",
			number: "0000000000000000",
			want:   true, // All zeros actually passes Luhn
		},
		{
			name:   "invalid/all_ones",
			number: "1111111111111111",
			want:   false,
		},
		{
			name:   "valid/single_digit",
			number: "0",
			want:   true,
		},
		{
			name:   "valid/two_digits",
			number: "18",
			want:   true,
		},
		{
			name:   "invalid/two_digits",
			number: "19",
			want:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := luhnValid(tt.number)
			assert.Equal(t, tt.want, got)
		})
	}
}
