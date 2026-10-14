package provider

import "github.com/jessola/better-ioc/internal/collection"

type Provider struct {
	services *collection.Collection
}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) Get[T any]() (v any, ok bool) {
	res, ok := p.services.Last[T]()
	if !ok {
		return
	}
	_ = res
	return
}
