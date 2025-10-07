package simapp

import (
	orbiter "github.com/noble-assets/orbiter"
)

func (app *SimApp) RegisterOrbiterControllers() {
	in := orbiter.ComponentsInputs{
		// Orbiter
		Orbiters: app.OrbiterKeeper,

		// Cosmos
		BankKeeper: app.BankKeeper,

		// Circle
		CCTPKeeper: app.CCTPKeeper,

		// Hyperlane
		HyperlaneCoreKeeper: app.HyperlaneKeeper,
		WarpKeeper:          &app.WarpKeeper,
	}

	orbiter.InjectComponents(in)
}
