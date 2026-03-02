// Binary policy consolidates the Guardrails and Data Governance services into
// a single gRPC server. These are stateless policy evaluation services with no
// inter-service dependencies.
package main

import (
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/boot"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/governance"
	"github.com/achyuthnsamudrala/bulkhead/control-plane/internal/guardrails"
	governancepb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/governance/v1"
	guardrailspb "github.com/achyuthnsamudrala/bulkhead/control-plane/pkg/gen/guardrails/v1"
)

func main() {
	f := boot.New(boot.Options{ServiceName: "bulkhead-policy"})

	// -----------------------------------------------------------------------
	// 1. Guardrails (requires DB for rule/set storage)
	// -----------------------------------------------------------------------
	guardrailsRepo := guardrails.NewPostgresRepository(f.DB)
	guardrailsSvc := guardrails.NewService(guardrailsRepo)
	guardrailsHandler := guardrails.NewHandler(guardrailsSvc)
	guardrailspb.RegisterGuardrailsServiceServer(f.Server, guardrailsHandler)

	// -----------------------------------------------------------------------
	// 2. Data Governance (stateless — no DB)
	// -----------------------------------------------------------------------
	governanceSvc := governance.NewService()
	governanceHandler := governance.NewHandler(governanceSvc)
	governancepb.RegisterDataGovernanceServiceServer(f.Server, governanceHandler)

	// -----------------------------------------------------------------------
	// Serve
	// -----------------------------------------------------------------------
	f.Logger.Info("registered services: Guardrails, DataGovernance")
	f.Serve()
}
