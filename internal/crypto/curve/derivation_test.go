package curve

import (
	"reflect"
	"testing"
	"strconv"
)

type args struct {
	pub  Key
	priv Key
}

type args3 struct{
	idx uint64
	der Key
	base Key
}

type TestCase struct {
	name    string
	args    args
	want    Key
	wantErr bool
}

type TestCase3 struct {
	name    string
	args    args3
	want    Key
	wantErr bool
}

type TestVectorDerive struct{
	Result string
	FirstArg string
	SecondArg string
}

type TestVectorGenerate struct{
	Result string
	FirstArg string
	SecondArg string
}

type TestToPublicKey struct {
	Result string
	FirstArg string
	SecondArg string
	ThirdArg string
}

type TestToPrivateKey struct {
	Result string
	FirstArg string
	SecondArg string
	ThirdArg string
}

// Test vectors:
var (
	deriveTestsPositive = []TestVectorDerive{TestVectorDerive{ 
	Result: "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",
	FirstArg: "1bbec189ed9e0aa5ed3f92686dbd17c7e76b96e4ebad67daf25c7a674193ee70",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	},TestVectorDerive{ Result: "5e6c5d86c47350484f801edfd3b1f4e96e135dd8f33f261fef54f6ef6051ea39",
	FirstArg: "506eb9f96a0e959d8ba5ddb68ad9ab460499b371eb95830137f585c54fd8fa09",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	},TestVectorDerive{ Result: "5e6c5d86c47350484f801edfd3b1f4e96e135dd8f33f261fef54f6ef6051ea39",
	FirstArg: "506eb9f96a0e959d8ba5ddb68ad9ab460499b371eb95830137f585c54fd8fa09",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "1dcdabed5ae341f9e978794e8049793ba054d2b001f8c6cd94bf08f209955223",
	FirstArg: "1a1ca7d7e74037e4d000a0fc2cc61389ac7d8b0a6b600c62e77374477c4c414d",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "f7f13fd9f261cc3dfba1e00b02b04ffc943ebd8993b05e58fc7ce9b798de96d2",
	FirstArg: "6cb53ddaa7575381195add2a8516b124009cc7374baecbf5220fab8e21db604d",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "cc152dadd7fee07b913a56a62e9f1a54c96ce2525de34de2a67a617e48353531",
	FirstArg: "3cdcc706f2714ab43a43e31fe8f187d8290663fb5f2ba4144a935f2957e10e87",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "d897fa2c0752ca058eca31056d6a4670f81c112306e927efde87d7576214c136",
	FirstArg: "f522d068e2d06190f18db8bedf337a7227ac4f4ecafde75ddb2414cb3c438ab6",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "79d2be74e27917a43eeae942f1fc630365e66d7ddf5e5d97738f3072922b7fab",
	FirstArg: "cf2ff5b2fd0452458ae053ae3e30f664cddb84d653de0a91f2a80a6db32a6714",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	},TestVectorDerive{ Result: "6a5d70d69c7471b447f93f41389dce6460b0f1ae21d807e2469677e31fbeda3d",
	FirstArg: "2552dea412bf0ccb14c122aeea2bce3c4540eb3a9bf2d9963ce37f1797e940b4",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "701003f729a4d6633a762589a002b9f16d5d1ad2d3acd60723e6a774c83bb86f",
	FirstArg: "a2a2993a6a6270ab5d3cbcfbefa6025f752ce97362c6764db98d64750e16cdd9",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "11491c4cf78e1c3c09dfc7278810391fbaf3f8b6c9e3af60250c0328928b2dbc",
	FirstArg: "1b38c7a007c50618315b39ea022dcd4a8ec99e504c92d2646ed1ae4ab4744911",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "d44add5f2ca28afddfe54d70d2787adbf23e9613567b6ed4b233765256f77545",
	FirstArg: "c9114c830d2d6a0d82dcc801dfc0f976ea4d8e6b46eb317bc7810876bd36aa69",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "06967f15b5ca897518cb42fbb529aec1a85f21910b17ead031808cd30a82b09c",
	FirstArg: "dfd3c039844c2dcc2e25ce5d8568fbb74910a9680c4354216a54b150079b51f6",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "79b072e96468609d52a8342b8f7b65f5382a933baf7a584e9df0e569ff3ece61",
	FirstArg: "ea1a8fa61900c5e69e6b74b0ab3518448a7cde6ce7b67eddd0505e6d8ba1a2b2",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "7e5fec7aca2dd238dfef40cfb1018587f650153609f193359211359c0d468926",
	FirstArg: "4e6ef4be3821b22e2f98ac08b5ef1ea4755fdbb4d38062b42fb3219e6750924c",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	},TestVectorDerive{ Result: "183c313584af62dc5ae5f73b1f1b4117709dfadd814567434517348aa048ef06",
	FirstArg: "78c7cd663b69bc21d79c6b2bb369c2f57075ee094574b9a127d4e70cf4eebb93",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "784f49fd35e562a359a65e3a6c384fd006eb7cb3f88d52c7104c6a838affb473",
	FirstArg: "8a7ec377a5a3c580091d87b4a1c16c949169255ec34793cafe554e758141ecf0",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "08deadba234ab89e8ebfad676c1875ffced8b9c2843359eac659d19aa09d1ae0",
	FirstArg: "41e52f6403fbb569804de00b3e06b381544d985004cde9085f4ff20d42aa950c",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "554c3af2407a0dec5b699ead5858866a34c88a7ec3b091ba96627612bd59ab41",
	FirstArg: "70481344b2dbd143422f375d328f545ba098c0e98c8818340f260fb5e2e4a07a",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "c77134a6d5fe89d670469ab9b769430fdf8c318314cab229fa1ca3815cdc3b25",
	FirstArg: "0cfde5c761aaecf7ea535c722a33ff67aa7877be42ee9d96bc91f1f0cebf2135",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "5ffd150ead2ae6c63addd49bb31aa0d3436f9d9e4b637c6995e1265603c72270",
	FirstArg: "139df9d50b930a9c002ce7bc20b07cbb3931960211536dfa3647d0fa257a9405",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "ad4306a21f90f9ab4f51e57cdddb59d6f9ee916416cbff5d7ca1a491181384bf",
	FirstArg: "0a7dd54a13fae750dfcc20b24bc6fe7059e4502837b0a3908a6a3cf6e3673414",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "d58a1ee317121a18ccc2ba4128a3c6b41150c90ed372c85a56abf0e6ff501776",
	FirstArg: "77be32e1be6b2794d6adb7694d1cd5ed2ce737b3ae0142210efb97574614fa9f",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "a7916de9a05437de5e65544632adce95c4f7477c7d029aba34dbe092c23c4342",
	FirstArg: "f8e585f1c1f1980774cfb74ea7326eedb9149b1ccdde39a97cb2c2c183e0d6a5",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}, TestVectorDerive{ Result: "98d12b76d56ba4043aaec1cd5eee6c982a9fe4e52f59154dba9d9249b4094d45",
	FirstArg: "87faf95af096be7e8f84ff6dafbde4a88333abc4fac30894779a899d1f9adca4",
	SecondArg: "9fde8d863a3040ff67ccc07c49b55ee4746d4db410fb18bdde7dbd7ccba4180e",
	}}
	generateTestPositive=[]TestVectorGenerate{
	TestVectorGenerate{Result:  "2a7c726e9a82008c5b9644bb55754249a3a5b6b258a811bd2ddd29ca41483f91",
	FirstArg:  "4e23535db80bb56d054d830db8713d5ac0b5f50b133e6ccee8b3a36cd310fe8d",
	SecondArg:  "ae4495fb374b8497484384dbc21a235676eaeae31207500a5b6a9ad5fb30e004",},
	TestVectorGenerate{Result:  "ec319ac875d25c4b46d16291ab27e8534b5a93fd9c5dd6b05f1de90809ecea50",
	FirstArg:  "2314952aa972dfe44b87f328215c25dc29704373a8c9e0378b8323ab7efa2332",
	SecondArg:  "4c0aa64894daddd3c63eb714e66df0884198015b71db0a1e70e3f63ec5f06308",},
	TestVectorGenerate{Result:  "fa50b03e0be626efcb160d188f28acd962340a7197a4d33f9e9f894312b06ce7",
	FirstArg:  "df0c8e98724d1a32b6cc737f1ad8cab83126e9f44a26eb27c6b005ddd33f5203",
	SecondArg:  "f46025a067b0565b7a1c1faa542625801a078d699ab24a06c856dfc92e4bcc02",},
	TestVectorGenerate{Result:  "526cf385d0b860b422dffca1143f8436b906514a4620c2182f402d8043d531fe",
	FirstArg:  "0b80e817435cc1d12fceef06cb6b86454bf310c6e5a746cf90d799a51a8e6a1a",
	SecondArg:  "8aa502c7c4ca8796282d0e9f6a31d9fec5f5c8db97fb7f7c2dd27893f46b1508",},
	TestVectorGenerate{Result:  "e88fb41909a49881053c0a3ab65156ab3f90b38fcefea111b1c89f4704866a6d",
	FirstArg:  "99f67df0f8fae6a7a66e316079fe6acc2697392c728fb9b5bfe9f038ef11ea66",
	SecondArg:  "6af4c823055ce361c386469c6856daa7b93fe1f7bf8c153a69b0609328db5601",},
	TestVectorGenerate{Result:  "1c7db86e17a6e1dee4f0eceb8bf84996aca4bda03bab41fa141739ff292ada2a",
	FirstArg:  "08c6a9ffa23961bea521eaa5de79903b3fc209476972654ff80fe626d23d41c6",
	SecondArg:  "e79797cb6581c2c50cd56a06a72d380a6a1a7696437ccc3f4f80973ee6fba006",},
	TestVectorGenerate{Result:  "a02f132ef0938e8d5a0c5ad2f5a3ee12527db96c02c721c7c3ac588c1a812695",
	FirstArg:  "9ebabf156826d2230119351ed88ed304df71257e727fbf5ea5cd3df15edda0ee",
	SecondArg:  "496827eeb9dd39700e6a58a5b1cf5af1c16f464d2b2f1858f33e023e73e4200c",},
	TestVectorGenerate{Result:  "2cf0c614a7d7012706aad4581f06b86d40cc0255a32b0789d782f955d385f3ec",
	FirstArg:  "04915c1b52b4005c9be24975a73abd96807c0446e837522e85fe06b9393beff5",
	SecondArg:  "aed39259c500aebfa61fa2b74f2ef8c1e635ee2ea36b9df570a0d3bce381f00a",},
	TestVectorGenerate{Result:  "a829535a3308fb11f2e09630b0274ef519e44a40078bd12c3351b76bc7c121f7",
	FirstArg:  "329423c43479104e51c526fb8837f446dfb9c0198a85248d1b81df5ed8b407aa",
	SecondArg:  "d804b121363630eda45cec6517672fa8ba12c0eaf5723c267e968277480aeb02",},
	TestVectorGenerate{Result:  "eb29de223892fb8565f9da69e50b981089bb62e24f7b573daf9c066ccc5254af",
	FirstArg:  "026b48cf8ef487bea48c6dc781befca433f03813a6954be5e1c52b3b6f8f1970",
	SecondArg:  "bb77ffb8e8155eb249527adcf64156facd2571179a6976672ce66a2e5f68730d",},
	TestVectorGenerate{Result:  "2cedaf5ad6f7a51828d7ec627874e5c27e6450be2cec3b4edc51fc35f2ca4cb2",
	FirstArg:  "ac2e98a88032c4b55c9e559c4ee8cd92aa05c251494782e81b93fa35c77add2c",
	SecondArg:  "2e7439f81fd1dae24f0ebc17f2db185054e6c2f6695731f09c5a1ea2726db508",},
	TestVectorGenerate{Result:  "ce6d93895e74dfe8ed708be5112675d89df5e295cd6ead1158136452b86f3cca",
	FirstArg:  "168350a40b4f93b6836bd4c0d662531b12912596490b8515cd336a39d2c27746",
	SecondArg:  "2d0d2f39746b6210ccb7d787586ab3ab0db3e9d0019262d6d58d52b98d45950d",},
	TestVectorGenerate{Result:  "bddbf0254c99fa94da948d88d8685bcee39508e867a0ffc391026b52280aa92e",
	FirstArg:  "dfcf268c1d86dce11ca10b01274b865e8aac6b16e1281b6a379001f149ba6809",
	SecondArg:  "70c32b506c11d9e8257429df4b4bd599a835b2b83ee9639ea21895c9d361af02",},
	TestVectorGenerate{Result:  "cea852fe7e63c4cbc6726d0dc4b2785d01c5606923a32eb7421f9407d3d3de27",
	FirstArg:  "cbd325d39ae6f18ccb737572611a00844e0063de8ebd3fae17dfca329f5e8aa0",
	SecondArg:  "105a2c962951031456aa79b8a21f30ebdba2757c5e75b2071102352ab74eed0a",},
	TestVectorGenerate{Result:  "b68c0871b7987633e79c46f5c9fbf99700b5605cbda71db3cfe93f683d7ee8f4",
	FirstArg:  "7d13b4eeec687ffc87adf4298e51d72eaf483dc37162c95b49a9d78a7634d2d8",
	SecondArg:  "679ef0ad7a97a17874d0f8074dab275e57279519344c2c3f288326590766870c",},
	TestVectorGenerate{Result:  "74709933a1e72aeedaf5c15108d79a49f435163e74cb0d27e0069029d1417ffc",
	FirstArg:  "ab47186ad08c6fe8dbc3521e85a3c95be505d3cacb51a7e8321a43ca3603ebb8",
	SecondArg:  "555b565dcd8c62797c70c3a2caf82929d01512be87cf69eee2358a2564c6f608",},
	TestVectorGenerate{Result:  "d8df1df309c242ed94275917f9862b244a2022824ff7fbfdf07119ca81b344fb",
	FirstArg:  "a2c4d295966716750191acbbcc1137ac67601a6218008add5c373e1f4f828a6a",
	SecondArg:  "fb90a61d2617c684268746a3e0e7fada7adc730bd500a4bd5152bf4048dcf30c",},
	TestVectorGenerate{Result:  "6c45d63934621d80055aab6c9ad8aa9388e1dd4c14eec1dc7385767dd5ef3106",
	FirstArg:  "dd55497995a4a5f217e7de8d2d7be40f31373134dd5b4ad6927abddb890b435e",
	SecondArg:  "cfb573410846f204f94f91e339169acd4d8c1eb8e7e174ea11772cb8df83d804",},
	TestVectorGenerate{Result:  "5544272b752f0a8287537aae51e303e2487c2be5b53d22d2fff436f40b7d6a97",
	FirstArg:  "aef4e0e912e9f25feedc17530d3b0bd25d09e283a020024352c50df54d4ad1c2",
	SecondArg:  "7b924b9bdd7c6147283f44581afb5ed67795f384afa39a9277def97a3797aa0b",},
	TestVectorGenerate{Result:  "efe5bd5b81e3e1f7c0e590205d78e2cecf7e5a63664259fa7d8133405db681e8",
	FirstArg:  "61dbe6050e187621255aa136d0e8db49ba68cd74fc158e010217735074139a6b",
	SecondArg:  "adb66f260d8a28b086f192d985f4cd5f838539b20910b5eb398930562827f505",},
	}
	ToPublicPositive = []TestToPublicKey{
	TestToPublicKey{Result: "4e23535db80bb56d054d830db8713d5ac0b5f50b133e6ccee8b3a36cd310fe8d",
	FirstArg: "4",
	SecondArg: "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",
	ThirdArg: "09917953e467c5cd62201ea63a93fcd123c754b249cb8e89d4251d67c907b169",},
	TestToPublicKey{Result: "2314952aa972dfe44b87f328215c25dc29704373a8c9e0378b8323ab7efa2332",
	FirstArg: "5",
	SecondArg: "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",
	ThirdArg: "09917953e467c5cd62201ea63a93fcd123c754b249cb8e89d4251d67c907b169",},
	TestToPublicKey{Result: "df0c8e98724d1a32b6cc737f1ad8cab83126e9f44a26eb27c6b005ddd33f5203",
	FirstArg: "8",
	SecondArg: "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",
	ThirdArg: "09917953e467c5cd62201ea63a93fcd123c754b249cb8e89d4251d67c907b169",},
	TestToPublicKey{Result: "0b80e817435cc1d12fceef06cb6b86454bf310c6e5a746cf90d799a51a8e6a1a",
	FirstArg: "9",
	SecondArg: "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",
	ThirdArg: "09917953e467c5cd62201ea63a93fcd123c754b249cb8e89d4251d67c907b169",},
	TestToPublicKey{Result: "99f67df0f8fae6a7a66e316079fe6acc2697392c728fb9b5bfe9f038ef11ea66",
	FirstArg: "10",
	SecondArg: "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",
	ThirdArg: "09917953e467c5cd62201ea63a93fcd123c754b249cb8e89d4251d67c907b169",},
	TestToPublicKey{Result: "08c6a9ffa23961bea521eaa5de79903b3fc209476972654ff80fe626d23d41c6",
	FirstArg: "12",
	SecondArg: "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",
	ThirdArg: "09917953e467c5cd62201ea63a93fcd123c754b249cb8e89d4251d67c907b169",},
	TestToPublicKey{Result: "9ebabf156826d2230119351ed88ed304df71257e727fbf5ea5cd3df15edda0ee",
	FirstArg: "13",
	SecondArg: "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",
	ThirdArg: "09917953e467c5cd62201ea63a93fcd123c754b249cb8e89d4251d67c907b169",},
	TestToPublicKey{Result: "04915c1b52b4005c9be24975a73abd96807c0446e837522e85fe06b9393beff5",
	FirstArg: "14",
	SecondArg: "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",
	ThirdArg: "09917953e467c5cd62201ea63a93fcd123c754b249cb8e89d4251d67c907b169",},
	}
	ToPrivatePositive = []TestToPrivateKey{
	TestToPrivateKey{Result:  "ae4495fb374b8497484384dbc21a235676eaeae31207500a5b6a9ad5fb30e004",
	FirstArg:  "4",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "4c0aa64894daddd3c63eb714e66df0884198015b71db0a1e70e3f63ec5f06308",
	FirstArg:  "5",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "f46025a067b0565b7a1c1faa542625801a078d699ab24a06c856dfc92e4bcc02",
	FirstArg:  "8",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "8aa502c7c4ca8796282d0e9f6a31d9fec5f5c8db97fb7f7c2dd27893f46b1508",
	FirstArg:  "9",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "6af4c823055ce361c386469c6856daa7b93fe1f7bf8c153a69b0609328db5601",
	FirstArg:  "10",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "e79797cb6581c2c50cd56a06a72d380a6a1a7696437ccc3f4f80973ee6fba006",
	FirstArg:  "12",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "496827eeb9dd39700e6a58a5b1cf5af1c16f464d2b2f1858f33e023e73e4200c",
	FirstArg:  "13",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "aed39259c500aebfa61fa2b74f2ef8c1e635ee2ea36b9df570a0d3bce381f00a",
	FirstArg:  "14",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "d804b121363630eda45cec6517672fa8ba12c0eaf5723c267e968277480aeb02",
	FirstArg:  "16",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "bb77ffb8e8155eb249527adcf64156facd2571179a6976672ce66a2e5f68730d",
	FirstArg:  "18",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "2e7439f81fd1dae24f0ebc17f2db185054e6c2f6695731f09c5a1ea2726db508",
	FirstArg:  "21",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "884c0528ee055c0ce8b9ab6c2fe3eeb0e5d12ac6c4829c57d407bef1a6073bdf",},
	TestToPrivateKey{Result:  "2d0d2f39746b6210ccb7d787586ab3ab0db3e9d0019262d6d58d52b98d45950d",
	FirstArg:  "1",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "5e6c5d86c47350484f801edfd3b1f4e96e135dd8f33f261fef54f6ef6051ea39",},
	TestToPrivateKey{Result:  "70c32b506c11d9e8257429df4b4bd599a835b2b83ee9639ea21895c9d361af02",
	FirstArg:  "2",
	SecondArg:  "e6887bea1e8126e8160ceef01ec35c81dd3e86e9d0e7e3c47087c113731ae508",
	ThirdArg:  "5e6c5d86c47350484f801edfd3b1f4e96e135dd8f33f261fef54f6ef6051ea39",}}
)

func TestDeriveKey(t *testing.T) {
	var tests []TestCase
	for _, el := range deriveTestsPositive {
		tests = append(tests, TestCase{name: "Derivation Pass", args: args{pub: hexToKey(el.FirstArg),priv: hexToKey(el.SecondArg)}, want: hexToKey(el.Result), wantErr: false})
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeriveKey(&tt.args.pub, &tt.args.priv)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeriveKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("DeriveKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestGenerateImage(t *testing.T) {
	var tests []TestCase
	for _, el := range generateTestPositive {
		tests = append(tests, TestCase{name: "Generate Pass", args: args{pub: hexToKey(el.FirstArg),priv: hexToKey(el.SecondArg)}, want: hexToKey(el.Result), wantErr: false})
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KeyImage(&tt.args.pub, &tt.args.priv)
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("KeyImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDerivationToPublicKey(t *testing.T) {
	var tests []TestCase3
	for _, el := range ToPublicPositive {
		value, _ := strconv.ParseUint(el.FirstArg,10,64)
		tests = append(tests, TestCase3{name: "Derivation to Public Key Pass", args: args3{idx: value, der: hexToKey(el.SecondArg), base: hexToKey(el.ThirdArg)}, want: hexToKey(el.Result), wantErr: false})
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DerivationToPublicKey(tt.args.idx, &tt.args.der, &tt.args.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("DerivationToPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("DerivationToPublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestDerivationToPrivateKey(t *testing.T) {
	var tests []TestCase3
	for _, el := range ToPrivatePositive {
		value, _ := strconv.ParseUint(el.FirstArg,10,64)
		tests = append(tests, TestCase3{name: "Derivation to Private Key Pass", args: args3{idx: value, base: hexToKey(el.SecondArg), der: hexToKey(el.ThirdArg)}, want: hexToKey(el.Result), wantErr: false})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DerivationToPrivateKey(tt.args.idx, &tt.args.base,&tt.args.der)
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("DerivationToPublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
