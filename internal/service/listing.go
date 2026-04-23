package service

import "easydrop/internal/pkg/listing"

var serviceListBounds = listing.Bounds{
	DefaultPage: 1,
	DefaultSize: 20,
	MaxSize:     100,
}
