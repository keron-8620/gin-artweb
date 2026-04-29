package crontab

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewCron(t *testing.T) {
	t.Run("create cron with logger", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		c := NewCron(logger)

		assert.NotNil(t, c)
		assert.IsType(t, &cron.Cron{}, c)
	})

	t.Run("cron instance can start and stop", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		c := NewCron(logger)

		c.Start()
		time.Sleep(10 * time.Millisecond)
		c.Stop()
	})

	t.Run("cron can add and run job", func(t *testing.T) {
		core, _ := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		c := NewCron(logger)

		_, err := c.AddFunc("* * * * *", func() {})
		assert.NoError(t, err)

		c.Start()
		time.Sleep(50 * time.Millisecond)
		c.Stop()
	})
}

func TestValidateCronExpression(t *testing.T) {
	tests := []struct {
		name            string
		expr            string
		withSeconds     bool
		wantValid       bool
		wantErrContains string
	}{
		{
			name:        "valid standard cron (5 fields)",
			expr:        "0 0 * * *",
			withSeconds: false,
			wantValid:   true,
		},
		{
			name:        "valid seconds cron (6 fields)",
			expr:        "0 0 0 * * *",
			withSeconds: true,
			wantValid:   true,
		},
		{
			name:        "valid standard cron with ranges",
			expr:        "0 9-17 * * 1-5",
			withSeconds: false,
			wantValid:   true,
		},
		{
			name:        "valid seconds cron with ranges",
			expr:        "0 0 9-17 * * 1-5",
			withSeconds: true,
			wantValid:   true,
		},
		{
			name:        "valid cron with step values",
			expr:        "0 */30 * * *",
			withSeconds: false,
			wantValid:   true,
		},
		{
			name:        "valid seconds cron with step values",
			expr:        "0 */30 * * * *",
			withSeconds: true,
			wantValid:   true,
		},
		{
			name:        "valid cron with wildcards",
			expr:        "* * * * *",
			withSeconds: false,
			wantValid:   true,
		},
		{
			name:        "valid seconds cron with wildcards",
			expr:        "* * * * * *",
			withSeconds: true,
			wantValid:   true,
		},
		{
			name:        "valid cron with specific values",
			expr:        "30 2 15 3 *",
			withSeconds: false,
			wantValid:   true,
		},
		{
			name:        "valid seconds cron with specific values",
			expr:        "15 30 2 15 3 *",
			withSeconds: true,
			wantValid:   true,
		},
		{
			name:            "invalid cron - wrong number of fields for standard",
			expr:            "0 0 0 * * *",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "invalid cron - wrong number of fields for seconds",
			expr:            "0 0 * * *",
			withSeconds:     true,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "invalid cron - out of range minute",
			expr:            "70 * * * *",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "invalid cron - out of range hour",
			expr:            "0 25 * * *",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "invalid cron - out of range day",
			expr:            "0 0 32 * *",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "invalid cron - out of range month",
			expr:            "0 0 * 13 *",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "invalid cron - out of range weekday",
			expr:            "0 0 * * 8",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "invalid cron - out of range seconds",
			expr:            "60 0 0 * * *",
			withSeconds:     true,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "invalid cron - invalid characters",
			expr:            "0 0 * * invalid",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "empty expression",
			expr:            "",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "invalid step value",
			expr:            "0 */0 * * *",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "invalid range",
			expr:            "0 10-5 * * *",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
		{
			name:            "mixed invalid characters",
			expr:            "abc def * * *",
			withSeconds:     false,
			wantValid:       false,
			wantErrContains: "cron表达式不合法",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ValidateCronExpression(tt.expr, tt.withSeconds)

			if tt.wantValid {
				assert.True(t, valid)
				assert.NoError(t, err)
			} else {
				assert.False(t, valid)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}
