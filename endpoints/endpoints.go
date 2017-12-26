package endpoints

import (
	"github.com/jinzhu/gorm"
)

// NewEndpoints is just below the http handlers
func NewEndpoints(
	db *gorm.DB,
) *Endpoints {
	return &Endpoints{
		db: db,
	}
}

// Endpoints represents all endpoints of the http server
type Endpoints struct {
	db *gorm.DB
}
