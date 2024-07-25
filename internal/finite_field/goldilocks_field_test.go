package finite_field_test

import (
	"testing"

	finite_field "intmax2-node/internal/finite_field"

	"github.com/iden3/go-iden3-crypto/ffg"
)

func TestWriteFixedSizeBytes(t *testing.T) {
	type args struct {
		buf  *finite_field.Buffer
		data []byte
	}
	tests := []struct {
		name             string
		args             args
		expectedInnerLen int
	}{
		{
			name: "success",
			args: args{
				buf:  finite_field.NewBuffer(make([]ffg.Element, 100)),
				data: []byte("test"),
			},
			expectedInnerLen: 1,
		},
		{
			name: "error",
			args: args{
				buf:  finite_field.NewBuffer(make([]ffg.Element, 100)),
				data: make([]byte, 33),
			},
			expectedInnerLen: 9,
		},
		{
			name: "error",
			args: args{
				buf:  finite_field.NewBuffer(make([]ffg.Element, 100)),
				data: make([]byte, 31),
			},
			expectedInnerLen: 8,
		},
		{
			name: "error",
			args: args{
				buf:  finite_field.NewBuffer(make([]ffg.Element, 100)),
				data: make([]byte, 0),
			},
			expectedInnerLen: 0,
		},
		{
			name: "success",
			args: args{
				buf:  finite_field.NewBuffer(make([]ffg.Element, 100)),
				data: []byte("test"),
			},
			expectedInnerLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finite_field.WriteFixedSizeBytes(tt.args.buf, tt.args.data, len(tt.args.data))
			innerLen := len(tt.args.buf.Inner())
			if innerLen != tt.expectedInnerLen {
				t.Errorf("WriteFixedSizeBytes() innerLen = %v, expectedInnerLen %v", innerLen, tt.expectedInnerLen)
			}
		})
	}
}
