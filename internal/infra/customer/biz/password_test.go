package biz

import (
	"testing"
)

func TestGetPasswordCharTypes(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected int
	}{
		{
			name:     "空密码",
			password: "",
			expected: 0,
		},
		{
			name:     "只有小写字母",
			password: "abcdef",
			expected: 1,
		},
		{
			name:     "只有大写字母",
			password: "ABCDEF",
			expected: 1,
		},
		{
			name:     "只有数字",
			password: "123456",
			expected: 1,
		},
		{
			name:     "只有特殊字符",
			password: "!@#$%^",
			expected: 1,
		},
		{
			name:     "小写字母和大写字母",
			password: "AbcDef",
			expected: 2,
		},
		{
			name:     "小写字母和数字",
			password: "abc123",
			expected: 2,
		},
		{
			name:     "小写字母和特殊字符",
			password: "abc!@#",
			expected: 2,
		},
		{
			name:     "大写字母和数字",
			password: "ABC123",
			expected: 2,
		},
		{
			name:     "大写字母和特殊字符",
			password: "ABC!@#",
			expected: 2,
		},
		{
			name:     "数字和特殊字符",
			password: "123!@#",
			expected: 2,
		},
		{
			name:     "小写字母、大写字母和数字",
			password: "Abc123",
			expected: 3,
		},
		{
			name:     "小写字母、大写字母和特殊字符",
			password: "Abc!@#",
			expected: 3,
		},
		{
			name:     "小写字母、数字和特殊字符",
			password: "abc123!@#",
			expected: 3,
		},
		{
			name:     "大写字母、数字和特殊字符",
			password: "ABC123!@#",
			expected: 3,
		},
		{
			name:     "包含所有类型字符",
			password: "Abc123!@#",
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPasswordCharTypes(tt.password)
			if result != tt.expected {
				t.Errorf("getPasswordCharTypes() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected int
	}{
		{
			name:     "空密码",
			password: "",
			expected: StrengthVeryWeak,
		},
		{
			name:     "极弱密码（短且单一类型）",
			password: "abc",
			expected: StrengthVeryWeak,
		},
		{
			name:     "弱密码（较短且类型少）",
			password: "abc123",
			expected: StrengthWeak,
		},
		{
			name:     "中等密码（长度适中且类型较多）",
			password: "Abc123",
			expected: StrengthWeak,
		},
		{
			name:     "强密码（较长且类型多）",
			password: "Abc123!@#",
			expected: StrengthMedium,
		},
		{
			name:     "极强密码（很长且类型多）",
			password: "Abc123!@#$%^&*",
			expected: StrengthStrong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPasswordStrength(tt.password)
			if result != tt.expected {
				t.Errorf("GetPasswordStrength() = %v, want %v", result, tt.expected)
			}
		})
	}
}
