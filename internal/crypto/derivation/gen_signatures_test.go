package derivation

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestCreateSignatures(t *testing.T) {
	fmt.Println("TXID:", hex.EncodeToString([]byte("\370\373\037\357\005c\034\b\035y۪\372\061\003\032\336l\341+n\216.\310\365\002\206?q\216\000P")))
	fmt.Println("KImage:", hex.EncodeToString([]byte("Sn\220n\033\f\325\177( \235\177h\aR\267\237\215S\240*M\035\022\257\356^\021\016\301\241\033")))
	fmt.Println("KImageGo:", hex.EncodeToString([]byte("Sn\220n\033\014\325\177( \235\177h\007R\267\237\215S\240*M\035\022\257\356^\021\016\301\241\033")))
											
	prefixHash, _ := hex.DecodeString("4e6823185a8771ea63ca39e1aa4dbb7578a6fba3da566cea485c82b3c0efc115")
	kImage, _ := hex.DecodeString("0f0ffd85c0de4993b1c69e5b0c21a3811b49f6c0f4097a81e33448c1149afaef")
	pub1, _ := hex.DecodeString("cad1c5a19a24b36c12ba2812652309ba4234f04ca20812ef2d76405d7740afe2")
	pub2, _ := hex.DecodeString("4ff3063f234d5d7b4717b2c2b9e4c09966f69f1c1083099420e37baf84d8c9ba")
	sec, _ := hex.DecodeString("071ae9ca27b1072d58abfc39b4bf143d61845202748c25189525172e1d081b0e")
	realIndex := 0

	var keyImage Key
	copy(keyImage[:], kImage)

	privKey := new(Key)
	copy(privKey[:], sec)

	mixins := make([]Key, 2)
	copy(mixins[0][:], pub1)
	copy(mixins[1][:], pub2)

	sigs, _ := GenerateRingSignature(prefixHash, keyImage, mixins, privKey, realIndex)


	fmt.Println("Len sigs: ", len(sigs))
	fmt.Println(sigs[0]);
	fmt.Println(sigs[1]);
	t.Errorf("Locked balance mismatch ") 

	// c1, _ := hex.DecodeString("25edaec9d801496716a0cd9aeb1e5757d5718f8c76108d108a8a1a9f8e51070f")
	// r1, _ := hex.DecodeString("640e50129530644f64e09631c6a7082fab5e8a551ce32b5806e74091322c2f0c")

	// c2, _ := hex.DecodeString("27c93cd9b8c53b561c58007173393a8b859bc0e4ac2ba110488c6501adf17508")
	// r2, _ := hex.DecodeString("02f13377e9f2059c623393c2ab2ad3c8f500e87e97dd35b8b480d9aa7f839305")

}
