package application

import (
	d "nsl-graph/internal/repository/domain"
)

type NetService struct {
	netRepo d.NetRepository
}

type NetServiceInt interface {
	// Brand
	AddBrand(brand string) error
	// Brand
	GetBrands() []string
}

func NewNetService(netRepository d.NetRepository) NetServiceInt {
	return &NetService{netRepo: netRepository}
}

func (ns *NetService) AddBrand(brand string) error {
	return ns.netRepo.AddBrand(brand)
}

func (ns *NetService) GetBrands() []string {
	return ns.netRepo.GetBrands()
}
