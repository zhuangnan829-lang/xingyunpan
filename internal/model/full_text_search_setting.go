package model

// FullTextSearchSetting stores admin-managed full text search preferences.
type FullTextSearchSetting struct {
	BaseModel
	Enabled         bool   `gorm:"not null;default:false" json:"enabled"`
	MeiliEndpoint   string `gorm:"size:255;not null;default:'http://localhost:7700'" json:"meili_endpoint"`
	APIKey          string `gorm:"size:255" json:"api_key"`
	ResultPageSize  int    `gorm:"not null;default:5" json:"result_page_size"`
	AISearch        bool   `gorm:"not null;default:false" json:"ai_search"`
	TikaEndpoint    string `gorm:"size:255;not null;default:'http://localhost:9998'" json:"tika_endpoint"`
	Extensions      string `gorm:"type:text" json:"extensions"`
	MaxFileSize     int    `gorm:"not null;default:25" json:"max_file_size"`
	MaxFileSizeUnit string `gorm:"size:8;not null;default:'MB'" json:"max_file_size_unit"`
	ChunkSize       int    `gorm:"not null;default:2000" json:"chunk_size"`
	ChunkUnit       string `gorm:"size:8;not null;default:'B'" json:"chunk_unit"`
	IndexNotes      string `gorm:"type:text" json:"index_notes"`
}

// TableName specifies the DB table name.
func (FullTextSearchSetting) TableName() string {
	return "full_text_search_settings"
}
