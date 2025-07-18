package simapp

import (
	orbiter "orbiter.dev"
)

func (app *SimApp) RegisterOrbiterControllers() {
	in := orbiter.ComponentsInputs{
		Orbiters: app.OrbiterKeeper,
	}

	orbiter.InjectComponents(in)
}
