package simapp

import (
	orbiter "github.com/noble-assets/orbiter"
)

func (app *SimApp) RegisterOrbiterControllers() {
	in := orbiter.ComponentsInputs{
		Orbiters:   app.OrbiterKeeper,
		BankKeeper: app.BankKeeper,
		CCTPKeeper: app.CCTPKeeper,
	}

	orbiter.InjectComponents(in)
}
