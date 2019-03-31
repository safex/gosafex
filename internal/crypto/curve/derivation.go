package curve

// DeriveKey derives a new private key derivation from a given public key
// and a secret.
func DeriveKey(pub Key, priv Key) Key {
	var point ExtendedGroupElement
	var point2 ProjectiveGroupElement
	var point3 CompletedGroupElement

	if !ScCheck(&priv) {
		panic("Invalid private key.")
	}
	tmp := pub
	if !point.fromBytes(&tmp) {
		panic("Invalid public key.")
	}

	tmp = priv
	GeScalarMult(&point2, &tmp, &point)
	GeMul8(&point3, &point2)
	point3.toProjective(&point2)

	point2.toBytes(&tmp)
	return tmp
}

func DerivationToPublicKey(idx uint64, derivation Key, baseKey Key) Key {
	var point1, point2 ExtendedGroupElement
	var point3 CachedGroupElement
	var point4 CompletedGroupElement
	var point5 ProjectiveGroupElement

	tmp := baseKey
	if !point1.fromBytes(&tmp) {
		panic("Invalid public key.")
	}
	scalar := KeyDerivationToScalar(idx, derivation)
	GeScalarMultBase(&point2, scalar)
	point2.toCached(&point3)
	geAdd(&point4, &point1, &point3)
	point4.toProjective(&point5)
	point5.toBytes(&tmp)
	return tmp
}

func KeyDerivation_To_PrivateKey(outputIndex uint64, baseKey Key, kd Key) Key {
	scalar := KeyDerivationToScalar(outputIndex, kd)

	tmp := baseKey
	ScAdd(&tmp, &tmp, scalar)
	return tmp
}

func GenerateKeyImage(pub Key, private Key) Key {
	var proj ProjectiveGroupElement

	ext := HashToEC(pub)
	GeScalarMult(&proj, &private, ext)

	var ki Key
	proj.toBytes(&ki)
	return ki
}

func HashToEC(data []byte) (result *ExtendedGroupElement) {
	result = new(ExtendedGroupElement)
	var p1 ProjectiveGroupElement
	var p2 CompletedGroupElement

	var h Key
	copy(h[:], data[:32])

	p1.fromBytes(&h)

	// fmt.Printf("p1 %+v\n", p1)
	GeMul8(&p2, &p1)
	p2.toExtended(result)
	return
}
