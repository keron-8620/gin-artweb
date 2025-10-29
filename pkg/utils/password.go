package utils

import (
	"unicode"
)

// Strength 等级常量
const (
	StrengthVeryWeak = iota
	StrengthWeak
	StrengthMedium
	StrengthStrong
	StrengthVeryStrong
)

// getPasswordCharTypes 分析密码包含的字符类型并返回类型数量
func getPasswordCharTypes(password string) (typesCount int) {
	if len(password) == 0 {
		return 0
	}

	// 使用位掩码标记字符类型
	const (
		maskLower   = 1 << iota // 1
		maskUpper               // 2
		maskDigit               // 4
		maskSpecial             // 8
	)

	var charTypes uint8

	// 遍历密码中的每个字符，识别其类型
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			charTypes |= maskLower
		case unicode.IsUpper(r):
			charTypes |= maskUpper
		case unicode.IsDigit(r):
			charTypes |= maskDigit
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			charTypes |= maskSpecial
		}
	}

	// 统计满足的字符类型数量
	if charTypes&maskLower != 0 {
		typesCount++
	}
	if charTypes&maskUpper != 0 {
		typesCount++
	}
	if charTypes&maskDigit != 0 {
		typesCount++
	}
	if charTypes&maskSpecial != 0 {
		typesCount++
	}

	return typesCount
}

// GetPasswordStrength 根据密码字符串返回其强度等级
func GetPasswordStrength(password string) int {
	passwordLen := len(password)
	
	// 空密码为极弱
	if passwordLen == 0 {
		return StrengthVeryWeak
	}

	// 获取密码包含的字符类型数量
	typesCount := getPasswordCharTypes(password)

	// 基于长度和字符类型数量计算强度分数
	score := passwordLen + typesCount*2 // 长度和类型多样性加权

	// 根据分数判断强度等级
	switch {
	case score < 10:
		return StrengthVeryWeak
	case score < 15:
		return StrengthWeak
	case score < 20:
		return StrengthMedium
	case score < 25:
		return StrengthStrong
	default:
		return StrengthVeryStrong
	}
}