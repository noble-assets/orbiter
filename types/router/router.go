// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2025, NASD Inc. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN "AS IS" BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package router

import (
	"errors"

	"orbiter.dev/types/interfaces"
)

// Router defines a generic router implementing the Router interface.
type Router[ID interfaces.IdentifierConstraint, T interfaces.Routable[ID]] struct {
	routes map[ID]T
	sealed bool
}

func New[ID interfaces.IdentifierConstraint, T interfaces.Routable[ID]]() interfaces.Router[ID, T] {
	return &Router[ID, T]{
		routes: make(map[ID]T),
	}
}

// Seal marks the router as sealed, preventing further route additions.
func (r *Router[ID, T]) Seal() {
	r.sealed = true
}

// Sealed returns whether the router is sealed.
func (r *Router[ID, T]) Sealed() bool {
	return r.sealed
}

// AddRoute adds a route to the router if it's not sealed.
func (r *Router[ID, T]) AddRoute(route T) error {
	if r.sealed {
		return errors.New("cannot add route to sealed router")
	}
	routeID := route.ID()
	if err := routeID.Validate(); err != nil {
		return errors.New("route id is not valid")
	}

	if r.HasRoute(routeID) {
		return errors.New("route is already set")
	}
	r.routes[routeID] = route

	return nil
}

// HasRoute checks if a route with the given ID exists.
func (r *Router[ID, T]) HasRoute(id ID) bool {
	_, exists := r.routes[id]
	return exists
}

// Route retrieves a route by ID.
func (r *Router[ID, T]) Route(id ID) (T, bool) {
	route, exists := r.routes[id]
	return route, exists
}
