package interfaces

import "orbiter.dev/types"

type IdentifierConstraint interface {
	types.ProtocolID | types.ActionID
	Validate() error
	String() string
}

// Routable defines the behavior required from a component
// to be used in a router.
type Routable[ID IdentifierConstraint] interface {
	// Returns the component's identifier.
	ID() ID
}

// RouterProvider defines the behavior required from a sub-keeper
// to manage accesses to a router.
type RouterProvider[ID IdentifierConstraint, T Routable[ID]] interface {
	Router() *Router[ID, T]
	SetRouter(*Router[ID, T])
}

type Router[ID IdentifierConstraint, T Routable[ID]] interface {
	Seal()
	Sealed() bool
	AddRoute(T)
	HasRoute(ID) bool
	Route(ID) (T, bool)
}
