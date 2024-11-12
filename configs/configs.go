package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/caarlos0/env/v8"
	"github.com/joho/godotenv"
)

const (
	hostPortDelimiter = ":"
	maxCORSMaxAge     = 600
)

type Config struct {
	APP                   APP
	API                   Api
	GRPC                  GRPC
	HTTP                  HTTP
	LOG                   LOG
	Wallet                Wallet
	PoW                   PoW
	Worker                Worker
	DepositSynchronizer   DepositSynchronizer
	BlockPostService      BlockPostService
	BlockValidityProver   BlockValidityProver
	BalanceValidityProver BalanceValidityProver
	WithdrawalService     WithdrawalService
	GasPriceOracle        GasPriceOracle
	Blockchain            Blockchain
	Network               Network
	StunServer            StunServer
	Swagger               Swagger
	OpenTelemetry         OpenTelemetry
	SQLDb                 SQLDb
	L2BatchIndex          L2BatchIndex
}

var once sync.Once
var config Config

func New() *Config {
	const intValue0 = 0
	_ = LoadDotEnv(intValue0)
	once.Do(func() {
		if err := env.Parse(&config); err != nil {
			const msg = "parsing configuration:"
			fmt.Println(msg, err)
			os.Exit(-1)
		}
		if config.HTTP.CORSMaxAge > maxCORSMaxAge {
			config.HTTP.CORSMaxAge = maxCORSMaxAge
		}
		config.HTTP.Timeout = config.HTTP.CORSMaxAge
		config.Swagger.Prepare()
		if config.GasPriceOracle.Delimiter <= intValue0 {
			config.GasPriceOracle.Delimiter = gasPriceOracleDelimiterDef
		}
		if config.GasPriceOracle.ExtraFee < intValue0 {
			config.GasPriceOracle.ExtraFee = intValue0
		}
		if config.BlockValidityProver.BlockValidityProverMaxValueOfInputDepositsInRequest <= intValue0 ||
			config.BlockValidityProver.BlockValidityProverMaxValueOfInputDepositsInRequest > blockValidityProverMaxValueOfInputDepositsInRequest {
			config.BlockValidityProver.BlockValidityProverMaxValueOfInputDepositsInRequest = blockValidityProverMaxValueOfInputDepositsInRequest
		}
		if config.BlockValidityProver.BlockValidityProverMaxValueOfInputTxRootInRequest <= intValue0 ||
			config.BlockValidityProver.BlockValidityProverMaxValueOfInputTxRootInRequest > blockValidityProverMaxValueOfInputTxRootInRequest {
			config.BlockValidityProver.BlockValidityProverMaxValueOfInputTxRootInRequest = blockValidityProverMaxValueOfInputTxRootInRequest
		}
		if config.APP.PrintConfig {
			config.print()
		}
	})
	return &config
}

func (cfg *Config) print() {
	jsonConfig, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		const (
			msg  = "marshal config:"
			code = -1
		)
		fmt.Println(msg, err)
		os.Exit(code)
	}
	fmt.Println(string(jsonConfig))
}

func LoadDotEnv(stepsUp int) error {
	const (
		path = "../"
		file = ".env"
	)
	if err := godotenv.Load(strings.Repeat(path, stepsUp) + file); err != nil {
		return err
	}
	return nil
}
