package swagger

import "embed"

//go:embed block_builder/*
var FsSwaggerBlockBuilder embed.FS

//go:embed store_vault/*
var FsSwaggerStoreVault embed.FS

//go:embed withdrawal/*
var FsSwaggerWithdrawal embed.FS
