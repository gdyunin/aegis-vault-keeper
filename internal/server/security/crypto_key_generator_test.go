package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCryptoKeyGenerator(t *testing.T) {
	t.Parallel()

	ckg := NewCryptoKeyGenerator()
	require.NotNil(t, ckg)
}

func TestCryptoKeyGenerator_CryptoKeyGenerate(t *testing.T) {
	t.Parallel()

	ckg := NewCryptoKeyGenerator()

	type args struct {
		size int
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
		wantErr bool
	}{
		{
			name:    "valid_key_size_16",
			args:    args{size: 16},
			wantLen: 16,
			wantErr: false,
		},
		{
			name:    "valid_key_size_32",
			args:    args{size: 32},
			wantLen: 32,
			wantErr: false,
		},
		{
			name:    "valid_key_size_64",
			args:    args{size: 64},
			wantLen: 64,
			wantErr: false,
		},
		{
			name:    "valid_key_size_1",
			args:    args{size: 1},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "valid_key_size_256",
			args:    args{size: 256},
			wantLen: 256,
			wantErr: false,
		},
		{
			name:    "invalid_zero_size",
			args:    args{size: 0},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "invalid_negative_size",
			args:    args{size: -1},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "invalid_large_negative_size",
			args:    args{size: -100},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ckg.CryptoKeyGenerate(tt.args.size)
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), "invalid key size")
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func TestCryptoKeyGenerator_Randomness(t *testing.T) {
	t.Parallel()

	ckg := NewCryptoKeyGenerator()

	// Test that multiple calls generate different keys
	keySize := 32
	numKeys := 10
	generatedKeys := make([][]byte, numKeys)

	for i := range numKeys {
		key, err := ckg.CryptoKeyGenerate(keySize)
		require.NoError(t, err)
		require.Len(t, key, keySize)
		generatedKeys[i] = key
	}

	// Check that all keys are different
	for i := range numKeys {
		for j := i + 1; j < numKeys; j++ {
			assert.NotEqual(t, generatedKeys[i], generatedKeys[j],
				"Keys at indices %d and %d should be different", i, j)
		}
	}
}

func TestCryptoKeyGenerator_KeyEntropy(t *testing.T) {
	t.Parallel()

	ckg := NewCryptoKeyGenerator()

	// Test that generated keys have good entropy (not all zeros, not all same value)
	tests := []struct {
		name string
		size int
	}{
		{"small_key", 16},
		{"medium_key", 32},
		{"large_key", 64},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			key, err := ckg.CryptoKeyGenerate(tt.size)
			require.NoError(t, err)
			require.Len(t, key, tt.size)

			// Check that key is not all zeros
			allZeros := true
			for _, b := range key {
				if b != 0 {
					allZeros = false
					break
				}
			}
			assert.False(t, allZeros, "Generated key should not be all zeros")

			// Check that key is not all the same value
			if len(key) > 1 {
				firstByte := key[0]
				allSame := true
				for _, b := range key[1:] {
					if b != firstByte {
						allSame = false
						break
					}
				}
				assert.False(t, allSame, "Generated key should not have all bytes the same")
			}
		})
	}
}

func TestCryptoKeyGenerator_ConcurrentGeneration(t *testing.T) {
	t.Parallel()

	ckg := NewCryptoKeyGenerator()

	// Test concurrent key generation
	const numGoroutines = 10
	const keySize = 32

	results := make(chan []byte, numGoroutines)
	errors := make(chan error, numGoroutines)

	for range numGoroutines {
		go func() {
			key, err := ckg.CryptoKeyGenerate(keySize)
			if err != nil {
				errors <- err
				return
			}
			results <- key
		}()
	}

	keys := make([][]byte, 0, numGoroutines)
	for range numGoroutines {
		select {
		case key := <-results:
			require.Len(t, key, keySize)
			keys = append(keys, key)
		case err := <-errors:
			t.Fatalf("Unexpected error during concurrent generation: %v", err)
		}
	}

	// Verify all keys are different
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			assert.NotEqual(t, keys[i], keys[j],
				"Concurrently generated keys should be different")
		}
	}
}

// Benchmark key generation performance.
func BenchmarkCryptoKeyGenerator_CryptoKeyGenerate(b *testing.B) {
	ckg := NewCryptoKeyGenerator()

	tests := []struct {
		name string
		size int
	}{
		{"16_bytes", 16},
		{"32_bytes", 32},
		{"64_bytes", 64},
		{"128_bytes", 128},
		{"256_bytes", 256},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for range b.N {
				_, err := ckg.CryptoKeyGenerate(tt.size)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
