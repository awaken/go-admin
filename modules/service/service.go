// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package service

type Service interface {
	Name() string
}

type Generator func() (Service, error)

func Register(k string, gen Generator) {
	if _, ok := services[k]; ok {
		panic("service has been already registered")
	}
	services[k] = gen
}

func GetServices() List {
	l := make(List, len(services))
	for k, gen := range services {
		var err error
		if l[k], err = gen(); err != nil {
			panic("service initialize fail")
		}
	}
	return l
}

var services = make(Generators)

type Generators map[string]Generator

type List map[string]Service

func (g List) MustGet(k string) Service {
	if v, ok := g[k]; ok {
		return v
	}
	panic("service not found")
}

func (g List) Get(k string) Service {
	return g[k]
}

func (g List) Add(k string, service Service) {
	if _, ok := g[k]; ok {
		panic("service exist")
	}
	g[k] = service
}
