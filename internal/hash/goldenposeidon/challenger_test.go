package goldenposeidon_test

import (
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"testing"

	"github.com/iden3/go-iden3-crypto/ffg"
	"github.com/stretchr/testify/assert"
)

func TestChallenger(t *testing.T) {
	t.Parallel()

	challenger := intMaxGP.NewChallenger()

	uintSlice := []uint64{1, 2}
	feSlice := []ffg.Element{}
	for _, v := range uintSlice {
		feSlice = append(feSlice, *ffg.NewElementFromUint64(v))
	}

	challenger.ObserveElements(feSlice)

	// Change feSlice for testing
	for i := range uintSlice {
		feSlice[i].SetUint64(3)
	}

	actual := challenger.GetNChallenges(24)
	expectedUint64 := []uint64{
		5670853448596375331, 5172341047684956727, 13364015481592832688, 4840440727741630358,
		16073914926410643468, 13716685746781302004, 11209788157740309596, 17222278624453642754,
		12081755210995097235, 13432333405837198024, 3447377635149567760, 7870753349118678807,
		5569056096572184370, 17685693549788741919, 14211757039835334950, 11597449937374612371,
		10445077586421379918, 355812654072660666, 15051621692528510111, 9398983630168452481,
		14391289183834088295, 4140863537182786900, 7440228235201741466, 14653535585180228638,
	}
	expected := []*ffg.Element{}
	for _, v := range expectedUint64 {
		expected = append(expected, ffg.NewElementFromUint64(v))
	}

	assert.Equal(t, actual, expected)

	uintSlice = []uint64{1, 2, 3}
	feSlice = []ffg.Element{}
	for _, v := range uintSlice {
		feSlice = append(feSlice, *ffg.NewElementFromUint64(v))
	}

	challenger.ObserveElements(feSlice)

	actual = challenger.GetNChallenges(9)
	expectedUint64 = []uint64{
		7464568625199957785, 5508170105610037514, 17582773042097119746, 551980252821407527,
		9390291698653197504, 12642903284687646509, 5223210455162382270, 5219070241010600415,
		4409774378469156838,
	}
	expected = []*ffg.Element{}
	for _, v := range expectedUint64 {
		expected = append(expected, ffg.NewElementFromUint64(v))
	}

	assert.Equal(t, actual, expected)
}
