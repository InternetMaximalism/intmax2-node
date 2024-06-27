package goldenposeidon_test

import (
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	"testing"

	"github.com/stretchr/testify/assert"
)

const prime uint64 = 18446744069414584321

func TestHash(t *testing.T) {
	t.Parallel()

	b0 := uint64(0)
	b1 := uint64(1)
	bm1 := prime - 1
	bM := prime

	h := intMaxGP.Hash([intMaxGP.NROUNDSF]uint64{b0, b0, b0, b0, b0, b0, b0, b0},
		[intMaxGP.CAPLEN]uint64{b0, b0, b0, b0})
	assert.Equal(t,
		[intMaxGP.CAPLEN]uint64{
			4330397376401421145,
			14124799381142128323,
			8742572140681234676,
			14345658006221440202,
		}, h,
	)

	h = intMaxGP.Hash([intMaxGP.NROUNDSF]uint64{1, 2, b0, b0, b0, b0, b0, b0},
		[intMaxGP.CAPLEN]uint64{b0, b0, b0, b0})
	assert.Equal(t,
		[intMaxGP.CAPLEN]uint64{
			17222278624453642754,
			11209788157740309596,
			13716685746781302004,
			16073914926410643468,
		}, h,
	)

	h = intMaxGP.Hash([intMaxGP.NROUNDSF]uint64{b1, b1, b1, b1, b1, b1, b1, b1},
		[intMaxGP.CAPLEN]uint64{b1, b1, b1, b1})
	assert.Equal(t,
		[intMaxGP.CAPLEN]uint64{
			16428316519797902711,
			13351830238340666928,
			682362844289978626,
			12150588177266359240,
		}, h,
	)

	h = intMaxGP.Hash([intMaxGP.NROUNDSF]uint64{b1, b1, b1, b1, b1, b1, b1, b1},
		[intMaxGP.CAPLEN]uint64{b1, b1, b1, b1})
	assert.Equal(t,
		[intMaxGP.CAPLEN]uint64{
			16428316519797902711,
			13351830238340666928,
			682362844289978626,
			12150588177266359240,
		}, h,
	)

	h = intMaxGP.Hash(
		[intMaxGP.NROUNDSF]uint64{bm1, bm1, bm1, bm1, bm1, bm1, bm1, bm1},
		[intMaxGP.CAPLEN]uint64{bm1, bm1, bm1, bm1},
	)
	assert.Equal(t,
		[intMaxGP.CAPLEN]uint64{
			13691089994624172887,
			15662102337790434313,
			14940024623104903507,
			10772674582659927682,
		}, h,
	)

	h = intMaxGP.Hash([intMaxGP.NROUNDSF]uint64{bM, bM, bM, bM, bM, bM, bM, bM},
		[intMaxGP.CAPLEN]uint64{b0, b0, b0, b0})
	assert.Equal(t,
		[intMaxGP.CAPLEN]uint64{
			4330397376401421145,
			14124799381142128323,
			8742572140681234676,
			14345658006221440202,
		}, h,
	)

	h = intMaxGP.Hash([intMaxGP.NROUNDSF]uint64{
		uint64(923978),
		uint64(235763497586),
		uint64(9827635653498),
		uint64(112870),
		uint64(289273673480943876),
		uint64(230295874986745876),
		uint64(6254867324987),
		uint64(2087),
	}, [intMaxGP.CAPLEN]uint64{b0, b0, b0, b0})
	assert.Equal(t,
		[intMaxGP.CAPLEN]uint64{
			1892171027578617759,
			984732815927439256,
			7866041765487844082,
			8161503938059336191,
		}, h,
	)
}
