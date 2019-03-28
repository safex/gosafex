package account

// Account contains methods of the account wrapper.
// You can get an accounts:
// - Address
// - Public key (view, spend)
// - Private key (view, spend)
type Account interface {
	Address() *Address
	PublicViewKey() PublicKey
	PublicSpendKey() PublicKey
	PrivateViewKey() PrivateKey
	PrivateSpendKey() PrivateKey
}

// Store is a wrapper struct containing all account information.
type Store struct {
	address  *Address
	viewKey  PrivateKey
	spendKey PrivateKey
}

// NewStore constructs a new store with the given address,
// view/spend private keys and a mnemonic.
func NewStore(adr *Address, viewPriv, spendPriv PrivateKey) *Store {
	return &Store{
		address:  adr,
		viewKey:  viewPriv,
		spendKey: spendPriv,
	}
}

// AddressMaker is a type of function that returns an address from view/spend public keys.
type AddressMaker = func(viewPub, spendPub PublicKey) *Address

func addressMaker(testnet bool) AddressMaker {
	if testnet {
		return NewRegularTestnetAddress
	}
	return NewRegularMainnetAdress
}

// GenerateAccount will create a new mainnet account store using a randomly generated seed.
// If testnet is true it will generate a testnet account.
// The implementation relies on system entropy from '/dev/urandom' by default.
// View keys are derived from spend keys.
// Returns an error if private keys cannot be generated.
func GenerateAccount(isTestnet bool) (result *Store, err error) {
	keyset, err := GenerateKeySet()
	if err != nil {
		return nil, err
	}
	adr := addressMaker(isTestnet)(keyset.view.pub, keyset.spend.pub)
	result = NewStore(adr, keyset.view.priv, keyset.spend.priv)
	return
}

// FromSeed will create a new account store using a given seed.
// If testnet is true it will generate a testnet account
// View keys are derived from spend keys.
func FromSeed(seed Seed, isTestnet bool) *Store {
	keyset := KeySetFromSeed(seed)
	adr := addressMaker(isTestnet)(keyset.view.pub, keyset.spend.pub)
	return NewStore(adr, keyset.view.priv, keyset.spend.priv)
}

// FromMnemonic will create a new account store using a given mnemonic.
// If testnet is true it will generate a testnet account
// View keys are derived from spend keys.
// Returns an error if the mnemonic is invalid or cannot be parsed.
func FromMnemonic(mnemonic *Mnemonic, isTestnet bool) (result *Store, err error) {
	seed, err := mnemonic.ToKey()
	if err != nil {
		return nil, err
	}
	result = FromSeed(seed, isTestnet)
	return
}

// Address implements Account. It returns the account's address.
func (s *Store) Address() *Address { return s.address }

// PublicViewKey implements Account. It returns the account's public view key.
func (s *Store) PublicViewKey() PublicKey { return s.address.ViewKey }

// PublicSpendKey implements Account. It returns the account's public spend key.
func (s *Store) PublicSpendKey() PublicKey { return s.address.SpendKey }

// PrivateSpendKey implements Account. It returns the account's private spend key.
func (s *Store) PrivateSpendKey() PrivateKey { return s.spendKey }

// PrivateViewKey implements Account. It returns the account's private view key.
func (s *Store) PrivateViewKey() PrivateKey { return s.viewKey }
