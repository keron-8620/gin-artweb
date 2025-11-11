package common

type PKUri struct {
	// 主键，必填，必须大于0
	// Required: true
	// Minimum: 1
	PK uint32 `uri:"pk" binding:"required,gt=0"`
}
