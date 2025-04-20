package kloudlite

import (
	"github.com/kloudlite/operator/toolkit/operator"
	v1 "github.com/kloudlite/plugin-k3s-cluster/api/v1"
	"github.com/kloudlite/plugin-k3s-cluster/internal/controller"
	"github.com/kloudlite/plugin-k3s-cluster/internal/env"
)

func RegisterInto(mgr operator.Operator) {
	ev, err := env.LoadEnv()
	if err != nil {
		panic(err)
	}

	mgr.AddToSchemes(v1.AddToScheme)
	mgr.RegisterControllers(
		&controller.K3sClusterReconciler{Env: ev},
	)
}
