package curve

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestCreateSignatures(t *testing.T) {

	fmt.Println("selected: ", []byte("\247kP\356\004j\352Ds\324C\212\333.\237\311\373\230G\354<\373\355\214\335\223\006\272\213\335\341\212"))

	fmt.Println("OutKey:", []byte("\200\210\214\272S\252\210\222\005p\255\071\202\032\264{\327X\203\177\230_\313\304CMy\003\334x\367"))
	fmt.Println("TXID:", hex.EncodeToString([]byte("i\210(\344!b܋b\256\317\372\226rT\321\177B\020&\356Sj\017j\305\365\003\\\254\373\220")))

	fmt.Println("PrefixHash: ", []byte("\305-@\247i&8\211\234\035\307NQzΏ\224\254\n\272\004\357̟-\377\216\n\374\244\344!"))
	fmt.Println("KImage:", hex.EncodeToString([]byte("\232\034h\241\b\b\242G\006sQL\363\231\235̕@o\234\252\236\\\363\002~\345\345\377m\336d")))

	fmt.Println("pub1:", hex.EncodeToString([]byte("'\334\376\t\207\370\030\064\f\366\065\341,c4F\254\b\366B\266\314oP\274\244\311\020i\304p/")))
	fmt.Println("pub2:", hex.EncodeToString([]byte("S\bT7\246\066тA\t:&S\227\241Od\274y$\216\223\351\345t\343\032\315\b۶\005")))

	fmt.Println("r: ", hex.EncodeToString([]byte("\202\016 \314Mܗ\301j\200\070\274Ւ\374\233\223\261n\350\246\300\337,\037\070b \031\003\a\005")))
	fmt.Println("c: ", hex.EncodeToString([]byte("(\367}c\345\332\t\324\070\310;=\365\070\026\276>\254d\027\063\335\066\034Q\331+\264\303\371\033")))

	prefixHash := []byte("H\215Y\037?\217\377\371$?v2\254\200u*\214.j\030\253\331\307\033ɸM\023\327\371|l")
	kImage := []byte("\367!\353\314\002\327\004\356\211\365\344\201\031ݹ\240\312\337J\313\311ܭt=\325\367\303\313]\340%")
	pub1 := []byte("*\332q\324шNI\305\370\371M\375\207\306\317\006i`\016\371r/o\245\352\350\262h\020.\375")
	pub2 := []byte("\260\071\060o}\202\247|&\315\341\031,\213\320\"6\267\212\243\201\223\246 jN\343\vs\f}.")
	sec := []byte("0Ti$o\025\275\334ݒ\316[\357\277-\310CS\027YѦ\346J\004?)\326\320\027|\006")
	realIndex := 0

	fmt.Println("Pubkey1: ", hex.EncodeToString(pub1))
	fmt.Println("Pubkey2: ", hex.EncodeToString(pub2))

	// Fake random

	c1 := []byte("v\237B\370\375^\003e\025\334&C\226\322Y1\362]\351\005\206\277\025\001\364\236\323\300\351\210\366\001")
	r1 := []byte("/\237W\233Zy\311\301\021\221\220w\222\342c\271\363\217\207R\016\327r\373\312l㖟he\003")

	c2 := []byte("?\314֏\236F\214\022fQ\270h`!\365\063\342\035\215\222q\347ݭ\026\344\311hO\200L\b")
	r2 := []byte("\361\234\202s\017\201-\212\233a\315\020\231,!l\350\371\021\366\207\231\030N\027\316z+\331%*\005")
	// Sigs:
	// $24 = std::vector of length 2, capacity 2 = {{c = {
	// 	data = "v\237B\370\375^\003e\025\334&C\226\322Y1\362]\351\005\206\277\025\001\364\236\323\300\351\210\366\001"}, r = {
	// 	data = "/\237W\233Zy\311\301\021\221\220w\222\342c\271\363\217\207R\016\327r\373\312l㖟he\003"}}, {c = {
	// 	data = "?\314֏\236F\214\022fQ\270h`!\365\063\342\035\215\222q\347ݭ\026\344\311hO\200L\b"}, r = {
	// 	data = "\361\234\202s\017\201-\212\233a\315\020\231,!l\350\371\021\366\207\231\030N\027\316z+\331%*\005"}}}

	var keyImage Key
	copy(keyImage[:], kImage)

	privKey := new(Key)
	copy(privKey[:], sec)

	mixins := make([]Key, 2)
	copy(mixins[0][:], pub1)
	copy(mixins[1][:], pub2)

	sigs, _ := GenerateRingSignature(prefixHash, keyImage, mixins, privKey, realIndex)

	fmt.Println("First1:", sigs[0].C, " ", sigs[0].R)
	fmt.Println("First2: ", hex.EncodeToString(c1), " ", hex.EncodeToString(r1))
	fmt.Println("Second1:", sigs[1].C, " ", sigs[1].R)
	fmt.Println("Second2: ", hex.EncodeToString(c2), " ", hex.EncodeToString(r2))

	// fmt.Println("cpp h: ", hex.EncodeToString([]byte("gi_\364\375\333\371%kPז\002\363\025i\372\f\266\232\027\354ꗴf\n]\252\236\305\001")))
	// fmt.Println("Len sigs: ", len(sigs))
	//fmt.Println("C", hex.EncodeToString([]byte("ֈ\325ա\343'n\252\233\336H\017ˑK\026,;\253U]\215\266\336\322\344\367`\210\242")), hex.EncodeToString([]byte("n\264\260\070\023ݼqk~\364\331F̻\207\350$\360|\321\063\226\"O\305\006ŐGH\006")))
	// fmt.Println(sigs[0]);
	t.Errorf("Locked balance mismatch ")

	// c1, _ := hex.DecodeString("25edaec9d801496716a0cd9aeb1e5757d5718f8c76108d108a8a1a9f8e51070f")
	// r1, _ := hex.DecodeString("640e50129530644f64e09631c6a7082fab5e8a551ce32b5806e74091322c2f0c")

	// c2, _ := hex.DecodeString("27c93cd9b8c53b561c58007173393a8b859bc0e4ac2ba110488c6501adf17508")
	// r2, _ := hex.DecodeString("02f13377e9f2059c623393c2ab2ad3c8f500e87e97dd35b8b480d9aa7f839305")
}
