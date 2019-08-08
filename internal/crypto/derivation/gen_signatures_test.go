package derivation

import (
	"encoding/hex"
	"testing"
)

func TestCreateSignatures(t *testing.T) {
	prefixHash, _ := hex.DecodeString("4e6823185a8771ea63ca39e1aa4dbb7578a6fba3da566cea485c82b3c0efc115")
	kImage, _ := hex.DecodeString("0f0ffd85c0de4993b1c69e5b0c21a3811b49f6c0f4097a81e33448c1149afaef")
	pub1, _ := hex.DecodeString("cad1c5a19a24b36c12ba2812652309ba4234f04ca20812ef2d76405d7740afe2")
	pub2, _ := hex.DecodeString("4ff3063f234d5d7b4717b2c2b9e4c09966f69f1c1083099420e37baf84d8c9ba")
	sec, _ := hex.DecodeString("071ae9ca27b1072d58abfc39b4bf143d61845202748c25189525172e1d081b0e")
	realIndex := 0

	var keyImage [32]byte
	copy(keyImage[:], kImage)
	
	mixins := make([][32]byte, 2)
	copy(mixins[0][:], pub1)
	copy(mixins[1][:], pub1)

	sigs := 


	c1, _ := hex.DecodeString("25edaec9d801496716a0cd9aeb1e5757")
	r1, _ := hex.DecodeString("d5718f8c76108d108a8a1a9f8e51070f")

	c2, _ := hex.DecodeString("640e50129530644f64e09631c6a7082f")
	r2, _ := hex.DecodeString("ab5e8a551ce32b5806e74091322c2f0c")

	c3, _ := hex.DecodeString("27c93cd9b8c53b561c58007173393a8b")
	r3, _ := hex.DecodeString("859bc0e4ac2ba110488c6501adf17508")

	c4, _ := hex.DecodeString("02f13377e9f2059c623393c2ab2ad3c8")
	r4, _ := hex.DecodeString("f500e87e97dd35b8b480d9aa7f839305")

}
