package third_party

import (
	"embed"
)

//go:embed OpenAPI/block_builder_service/*
var OpenAPIBlockBuilder embed.FS

//go:embed OpenAPI/store_vault_service/*
var OpenAPIStoreVault embed.FS

//go:embed OpenAPI/withdrawal_service/*
var OpenAPIWithdrawal embed.FS

//go:embed OpenAPI/block_validity_prover_service/*
var OpenAPIBlockValidityProver embed.FS
