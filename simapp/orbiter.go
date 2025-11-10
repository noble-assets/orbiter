package simapp

import (
	orbiter "github.com/noble-assets/orbiter/v2"
)

func (app *SimApp) RegisterOrbiterControllers() {
	in := orbiter.ComponentsInputs{
		Orbiters:   app.OrbiterKeeper,
		BankKeeper: app.BankKeeper,
		CCTPKeeper: app.CCTPKeeper,
		WarpKeeper: app.WarpKeeper,
	}

	orbiter.InjectComponents(in)
}
