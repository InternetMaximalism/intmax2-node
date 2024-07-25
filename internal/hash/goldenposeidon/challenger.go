package goldenposeidon

/// This is the Go implementation of the following library.
/// https://github.com/0xPolygonZero/plonky2/blob/598ac876aea2cae5252ed3c2760bc441b7e8b661/plonky2/src/iop/challenger.rs

import (
	"github.com/iden3/go-iden3-crypto/ffg"
)

type Challenger struct {
	spongeState  [mLen]*ffg.Element
	inputBuffer  []*ffg.Element
	outputBuffer []*ffg.Element
}

func NewChallenger() *Challenger {
	spongeState := [mLen]*ffg.Element{}
	for i := 0; i < mLen; i++ {
		spongeState[i] = new(ffg.Element).SetUint64(0)
	}

	return &Challenger{
		spongeState:  spongeState,
		inputBuffer:  []*ffg.Element{},
		outputBuffer: []*ffg.Element{},
	}
}

func (c *Challenger) Reset() {
	c.inputBuffer = []*ffg.Element{}
	c.outputBuffer = []*ffg.Element{}
}

func (c *Challenger) ObserveElement(element ffg.Element) {
	// Any buffered outputs are now invalid, since they wouldn't reflect this input.
	c.outputBuffer = []*ffg.Element{}

	c.inputBuffer = append(c.inputBuffer, new(ffg.Element).Set(&element))

	if len(c.inputBuffer) == NROUNDSF {
		c.duplexing()
	}
}

func (c *Challenger) ObserveElements(elements []ffg.Element) {
	for _, element := range elements {
		c.ObserveElement(element)
	}
}

func (c *Challenger) GetChallenge() *ffg.Element {
	// If we have buffered inputs, we must perform a duplexing so that the challenge will
	// reflect them. Or if we've run out of outputs, we must perform a duplexing to get more.
	if len(c.inputBuffer) != 0 || len(c.outputBuffer) == 0 {
		c.duplexing()

		// debug assertion
		// if len(c.outputBuffer) == 0 {
		// 	panic("duplexing failed")
		// }
	}

	e := c.outputBuffer[len(c.outputBuffer)-1]
	c.outputBuffer = c.outputBuffer[:len(c.outputBuffer)-1]

	return new(ffg.Element).Set(e)
}

func (c *Challenger) GetNChallenges(n int) []*ffg.Element {
	challenges := []*ffg.Element{}
	for i := 0; i < n; i++ {
		e := c.GetChallenge()
		challenges = append(challenges, e)
	}

	return challenges
}

func (c *Challenger) duplexing() {
	// assert!(self.input_buffer.len() <= H::Permutation::RATE);
	if len(c.inputBuffer) > NROUNDSF {
		panic("input buffer too large")
	}

	// Overwrite the first r elements with the inputs. This differs from a standard sponge,
	// where we would xor or add in the inputs. This is a well-known variant, though,
	// sometimes called "overwrite mode".
	for i, e := range c.inputBuffer {
		c.spongeState[i].Set(e)
	}
	c.inputBuffer = []*ffg.Element{}

	// Apply the permutation.
	c.spongeState = Permute(c.spongeState)

	c.outputBuffer = []*ffg.Element{}

	// Squeeze out the first r elements to the output buffer.
	for i := 0; i < NROUNDSF; i++ {
		c.outputBuffer = append(c.outputBuffer, c.spongeState[i])
	}
}
