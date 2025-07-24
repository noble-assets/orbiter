package simapp

import (
	orbiter "orbiter.dev"
)

func (app *SimApp) RegisterOrbiterControllers() {
	in := orbiter.ComponentsInputs{
		Orbiters:   app.OrbiterKeeper,
		BankKeeper: app.BankKeeper,
		CCTPKeeper: app.CCTPKeeper,
	}

	orbiter.InjectComponents(in)
}
