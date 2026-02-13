package model

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type PackageModel struct {
	database.BaseModel
	Label           string    `gorm:"column:label;type:varchar(50);index:idx_package_label;comment:标签" json:"label"`
	StorageFilename string    `gorm:"column:storage_filename;type:varchar(50);not null;uniqueIndex;comment:磁盘存储文件名" json:"storage_filename"`
	OriginFilename  string    `gorm:"column:origin_filename;type:varchar(255);comment:原始文件名" json:"origin_filename"`
	Version         string    `gorm:"column:version;type:varchar(50);comment:版本号" json:"version"`
	UploadedAt      time.Time `gorm:"column:uploaded_at;autoCreateTime;comment:上传时间" json:"uploaded_at"`
}

func (m *PackageModel) TableName() string {
	return "resource_package"
}

func (m *PackageModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.BaseModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("label", m.Label)
	enc.AddString("storage_filename", m.StorageFilename)
	enc.AddString("origin_filename", m.OriginFilename)
	enc.AddString("version", m.Version)
	enc.AddTime("uploaded_at", m.UploadedAt)
	return nil
}
