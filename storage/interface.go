package storage

import "github.com/Lan-ce-lot/data-people/models"

// Storage 存储接口
type Storage interface {
	// Init 初始化存储
	Init() error

	// Save 保存单个文章
	Save(article *models.Article) error

	// SaveBatch 批量保存文章
	SaveBatch(articles []*models.Article) error

	// Close 关闭存储连接
	Close() error

	// GetStorageType 获取存储类型
	GetStorageType() string
}
