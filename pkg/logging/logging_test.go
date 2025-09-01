package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		level         string
		expectedLevel zapcore.Level
		shouldBeValid bool
	}{
		{
			name:          "debug level",
			level:         "debug",
			expectedLevel: zapcore.DebugLevel,
			shouldBeValid: true,
		},
		{
			name:          "info level",
			level:         "info",
			expectedLevel: zapcore.InfoLevel,
			shouldBeValid: true,
		},
		{
			name:          "warn level",
			level:         "warn",
			expectedLevel: zapcore.WarnLevel,
			shouldBeValid: true,
		},
		{
			name:          "error level",
			level:         "error",
			expectedLevel: zapcore.ErrorLevel,
			shouldBeValid: true,
		},
		{
			name:          "fatal level",
			level:         "fatal",
			expectedLevel: zapcore.FatalLevel,
			shouldBeValid: true,
		},
		{
			name:          "panic level",
			level:         "panic",
			expectedLevel: zapcore.PanicLevel,
			shouldBeValid: true,
		},
		{
			name:          "uppercase DEBUG",
			level:         "DEBUG",
			expectedLevel: zapcore.DebugLevel,
			shouldBeValid: true,
		},
		{
			name:          "uppercase INFO",
			level:         "INFO",
			expectedLevel: zapcore.InfoLevel,
			shouldBeValid: true,
		},
		{
			name:          "invalid level",
			level:         "invalid",
			expectedLevel: zapcore.InfoLevel, // defaults to INFO
			shouldBeValid: false,
		},
		{
			name:          "empty level",
			level:         "",
			expectedLevel: zapcore.InfoLevel, // defaults to INFO
			shouldBeValid: false,
		},
		{
			name:          "random string",
			level:         "randomlevel",
			expectedLevel: zapcore.InfoLevel, // defaults to INFO
			shouldBeValid: false,
		},
		{
			name:          "numeric string",
			level:         "123",
			expectedLevel: zapcore.InfoLevel, // defaults to INFO
			shouldBeValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := NewLogger(tt.level)

			// Verify logger is created
			require.NotNil(t, logger, "Logger should not be nil")
			assert.IsType(t, &zap.SugaredLogger{}, logger, "Should return SugaredLogger")

			// Verify logger is functional
			assert.NotPanics(t, func() {
				logger.Info("Test log message")
				logger.Debug("Test debug message")
				logger.Warn("Test warn message")
				logger.Error("Test error message")
			}, "Logger should be functional")
		})
	}
}

func TestNewLogger_LogLevels(t *testing.T) {
	t.Parallel()

	levelTests := []struct {
		name  string
		level string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
		{"fatal", "fatal"},
		{"panic", "panic"},
	}

	for _, tt := range levelTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := NewLogger(tt.level)

			// Test that logger can handle all methods without panicking
			assert.NotPanics(t, func() {
				logger.Debug("debug message")
				logger.Info("info message")
				logger.Warn("warn message")
				logger.Error("error message")

				// Test with format strings
				logger.Debugf("debug %s", "formatted")
				logger.Infof("info %s", "formatted")
				logger.Warnf("warn %s", "formatted")
				logger.Errorf("error %s", "formatted")

				// Test with key-value pairs
				logger.Debugw("debug with fields", "key", "value")
				logger.Infow("info with fields", "key", "value")
				logger.Warnw("warn with fields", "key", "value")
				logger.Errorw("error with fields", "key", "value")
			})
		})
	}
}

func TestNewLogger_InvalidLevels(t *testing.T) {
	t.Parallel()

	invalidLevels := []string{
		"invalid",
		"",
		"123",
		"trace", // not a valid zap level
		"verbose",
		"quiet",
		"off",
		"ALL",
		"   ", // whitespace
		"null",
		"undefined",
	}

	for _, level := range invalidLevels {
		t.Run("invalid_level_"+level, func(t *testing.T) {
			t.Parallel()

			// Should not panic even with invalid levels
			assert.NotPanics(t, func() {
				logger := NewLogger(level)
				require.NotNil(t, logger)

				// Should still be functional (defaults to INFO)
				logger.Info("Test message with invalid level")
			})
		})
	}
}

func TestNewLogger_ConsistentBehavior(t *testing.T) {
	t.Parallel()

	// Test that multiple calls with same level return functional loggers
	logger1 := NewLogger("info")
	logger2 := NewLogger("info")

	require.NotNil(t, logger1)
	require.NotNil(t, logger2)

	// Both should be functional
	assert.NotPanics(t, func() {
		logger1.Info("Message from logger1")
		logger2.Info("Message from logger2")
	})
}

func TestNewLogger_CaseSensitivity(t *testing.T) {
	t.Parallel()

	caseSensitiveTests := []struct {
		name  string
		level string
	}{
		{"lowercase", "debug"},
		{"uppercase", "DEBUG"},
		{"mixed case", "Debug"},
		{"weird case", "dEbUg"},
	}

	for _, tt := range caseSensitiveTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := NewLogger(tt.level)
			require.NotNil(t, logger)

			// Should not panic regardless of case
			assert.NotPanics(t, func() {
				logger.Debug("Debug message")
				logger.Info("Info message")
			})
		})
	}
}

func TestNewLogger_WithDifferentLevels(t *testing.T) {
	t.Parallel()

	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run("level_"+level, func(t *testing.T) {
			t.Parallel()

			logger := NewLogger(level)
			require.NotNil(t, logger)

			// Test message output at different levels
			assert.NotPanics(t, func() {
				logger.Debug("Debug message")
				logger.Info("Info message")
				logger.Warn("Warn message")
				logger.Error("Error message")
			})
		})
	}
}
