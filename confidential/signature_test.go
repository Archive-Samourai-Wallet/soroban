package confidential

import (
	"testing"
)

func Test_signMessage(t *testing.T) {
	type args struct {
		privateKey string
		message    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"test", args{"L3xJb1qTaa5DUpmMgb2yKMy9n1nxCYAPuMhA34EeZ3Ua2Xr9wyDF", "Hello, World!"}, "30440220046e86f0bff9639a893616e1db3abfa24cafa8818e7e47798c860d5982968ef502200241904a24128f6f73b8f5675368ff85992aa2b97bb40fe91ab361c96c62ca35"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := signMessage(tt.args.privateKey, tt.args.message); got != tt.want {
				t.Errorf("signMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_verifyEcdsaSignature(t *testing.T) {
	type args struct {
		publicKey string
		message   string
		signature string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test", args{"024d1d2028d6a503c5d688425eddcb9a348696d606fb6d521b8a336de760d51e8e", "Hello, World!", "30440220046e86f0bff9639a893616e1db3abfa24cafa8818e7e47798c860d5982968ef502200241904a24128f6f73b8f5675368ff85992aa2b97bb40fe91ab361c96c62ca35"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verifyEcdsaSignature(tt.args.publicKey, tt.args.message, tt.args.signature); got != tt.want {
				t.Errorf("verifyEcdsaSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
