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

func Test_verifyTestnet3Signature(t *testing.T) {
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
		{"failed", args{"mi42XN9J3eLdZae4tjQnJnVkCcNDRuAtz4", "hello_failed", "IOMVJ0SDwbDs1zb3IV/MxEeNRwn8FA+2ZZlmtE6HzGEeMxm2lSDNSHoJmNCCNghIPHAJxWg6smIrItgvzofllEg="}, false},
		{"verified", args{"mi42XN9J3eLdZae4tjQnJnVkCcNDRuAtz4", "hello", "IOMVJ0SDwbDs1zb3IV/MxEeNRwn8FA+2ZZlmtE6HzGEeMxm2lSDNSHoJmNCCNghIPHAJxWg6smIrItgvzofllEg="}, true},
		{"verified-message", args{"mi42XN9J3eLdZae4tjQnJnVkCcNDRuAtz4", "com.samourai.whirlpool.wo.1687247300669", "IMhtg2RvUjfy18BP0wYpi9CpQgrY2/DVPVTK3W/6vVGYeZyR+WQD4Kt3bEmC8eSvgpKy9X4f5nFzZu/CiJIS8sc="}, true},
		{"verified-core", args{"mk7YZqsP6jEJ4XNqdDQEXYwR4umRKceddR", "Test", "IIQCJwCvFQ62E7JlOsozLbyjybqLE719G1hPxZcJBANGIxP7rtv5Rg9RFJ3gsBe19kbeFyKfaKqFGUIGbuZHORI="}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verifyTestnet3Signature(tt.args.publicKey, tt.args.message, tt.args.signature); got != tt.want {
				t.Errorf("verifyTestnet3Signature() = %v, want %v", got, tt.want)
			}
		})
	}
}
