package ProfileForm

import "testing"

func TestKnownUnsupportedBFloat16Capability(t *testing.T) {
	tests := []struct {
		name              string
		precision         string
		computeCapability float32
		want              bool
	}{
		{name: "detection pending", precision: "bfloat16", computeCapability: 0, want: false},
		{name: "old CUDA GPU", precision: "bfloat16", computeCapability: 7.5, want: true},
		{name: "Ampere boundary", precision: "bfloat16", computeCapability: 8.0, want: false},
		{name: "Blackwell", precision: "bfloat16", computeCapability: 12.0, want: false},
		{name: "quantized bfloat16 on old GPU", precision: "int8_bfloat16", computeCapability: 7.5, want: true},
		{name: "other precision", precision: "float16", computeCapability: 7.5, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := hasKnownUnsupportedBFloat16Capability(test.precision, test.computeCapability)
			if got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}
