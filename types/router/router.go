package router

import "orbiter.dev/types/interfaces"

// Router defines a generic router implementing the Router interface.
type Router[ID interfaces.IdentifierConstraint, T interfaces.Routable[ID]] struct {
	routes map[ID]T
	sealed bool
}

func NewRouter[ID interfaces.IdentifierConstraint, T interfaces.Routable[ID]]() interfaces.Router[ID, T] {
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
func (r *Router[ID, T]) AddRoute(route T) {
	if r.sealed {
		panic("cannot add route to sealed router")
	}
	routeId := route.ID()
	if err := routeId.Validate(); err != nil {
		panic("route id is not valid")
	}

	if r.HasRoute(routeId) {
		panic("route is already set")
	}
	r.routes[route.ID()] = route
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
