package service

type PkUri struct {
	// 主键，必填，必须大于0
	// Required: true
	// Minimum: 1
	Pk uint32 `uri:"pk" binding:"required,gt=0"`
}
