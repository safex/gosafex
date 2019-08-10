package derivation

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestCreateSignatures(t *testing.T) {
	fmt.Println("PrefixHash: ", []byte("\001\000\001\002\200ॖ\273\021\002\031\000nЕ\367\350\243\335\331SQ\340,\352\321P9\202\271A\031ꧡ+\242\373\031\027\300(Y\330\005\200ғ\255\003\000\002\244\234\313\376`3\034!\377\271I_\255u\331S\330P>\f\361a\341x\205m\214\035\222\064\242\361\200\220\312\322\306\016\000\002\022\377\225\027\367\063\323|\035%\276\367\375T\023\347\365\346\310\336/\225\002\320V#ٽ\232\352\"\275\200\264\304\303!\000\002\022\251\306Sԯ\357\377\312O\334\362\362\354\302H\202)\317^0\355R\352x^\302\216\334\b\030Ӏ\220\337\300J\000\002\226$\270\064\066!\261\003\034y\256\364E\271\362\324<\333l\256\276\271~i\232-\330R\244"))
	fmt.Println("OutKey:", []byte("\200\210\214\272S\252\210\222\005p\255\071\202\032\264{\327X\203\177\230_\313\304CMy\003\334x\367"))
	fmt.Println("TXID:", hex.EncodeToString([]byte("i\210(\344!b܋b\256\317\372\226rT\321\177B\020&\356Sj\017j\305\365\003\\\254\373\220")))
	fmt.Println("KImage:", hex.EncodeToString([]byte("hNK\311\034\240\346d|VG\253\303\fK\233\066\214\034\207\333\327A\210\226gR\336\332\363Nq")))
	fmt.Println("KHImage:", hex.EncodeToString([]byte("n\320\225\367\350\243\335\331SQ\340,\352\321P9\202\271A\031\352\247\241+\242\373\031\027\300(Y\330")))
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
