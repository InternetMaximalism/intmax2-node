package balance_service_test

import (
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserBalance(t *testing.T) {
	userAddressHex := "0x030644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd3"

	userAddress, err := intMaxAcc.NewAddressFromHex(userAddressHex)
	assert.NoError(t, err)
	ethBalance, err := balance_service.GetUserBalance(userAddress, 0)
	assert.NoError(t, err)
	usdcBalance, err := balance_service.GetUserBalance(userAddress, 1)
	assert.NoError(t, err)
	wbtcBalance, err := balance_service.GetUserBalance(userAddress, 2)
	assert.NoError(t, err)

	fmt.Printf("ETH balance: %s\n", ethBalance)
	fmt.Printf("USDC balance: %s\n", usdcBalance)
	fmt.Printf("WBTC balance: %s\n", wbtcBalance)

}
