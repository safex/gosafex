package account

// Account contains methods of the account wrapper.
// You can get an accounts:
// - Address
// - Public key (view, spend)
// - Private key (view, spend)
type Account interface {
	Address() Address
	PublicViewKey() PublicKey
	PublicSpendKey() PublicKey
	PrivateViewKey() PrivateKey
	PrivateSpendKey() PrivateKey
}

// Store is a wrapper struct containing all account information.
type Store struct {
	address  Address
	viewKey  PrivateKey
	spendKey PrivateKey
}

// Address returns the account's address.
func (s *Store) Address() Address { return s.address }

// PublicViewKey returns the account's public view key.
func (s *Store) PublicViewKey() PublicKey { return s.address.ViewKey }

// PublicSpendKey returns the account's public spend key.
func (s *Store) PublicSpendKey() PublicKey { return s.address.SpendKey }

// PrivateSpendKey returns the account's private spend key.
func (s *Store) PrivateSpendKey() PrivateKey { return s.spendKey }

// PrivateViewKey returns the account's private view key.
func (s *Store) PrivateViewKey() PrivateKey { return s.viewKey }
