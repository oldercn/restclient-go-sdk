package domain

import (
	"encoding/hex"
)

// Lengths of hashes, identity and publicKeys in bytes.
const (
	//IdentityLength is the expected length of the address
	IdentityLength = 32
	// HashLength is the expected length of the hash
	HashLength = 32
	// PublicKeyLength is the expected length of the publicKey
	PublicKeyLength = 64
)

type Identity [IdentityLength]byte

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToIdentity(b []byte) Identity {
	var i Identity
	i.SetBytes(b)
	return i
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (id *Identity) SetBytes(b []byte) {
	if len(b) > len(id) {
		b = b[len(b)-IdentityLength:]
	}
	copy(id[IdentityLength-len(b):], b)
}

func (id *Identity) ToHex() string {
	return hex.EncodeToString(id[:])
}

// Bytes gets the string representation of the underlying address.
func (id *Identity) Bytes() []byte { return id[:] }

type Hash [HashLength]byte

// SetBytes sets the hash to the value of b.
// If b is larger than len(h), b will be cropped from the left.
func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

// BytesToHash sets b to hash.
// If b is larger than len(h), b will be cropped from the left.
func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func (h *Hash) ToHex() string {
	return hex.EncodeToString(h[:])
}

// Bytes gets the byte representation of the underlying hash.
func (h Hash) Bytes() []byte { return h[:] }

type PublicKey [PublicKeyLength]byte

func (p *PublicKey) ToHex() string {
	return hex.EncodeToString(p[:])
}

// SetBytes sets the hash to the value of b.
// If b is larger than len(h), b will be cropped from the left.
func (p *PublicKey) SetBytes(b []byte) {
	if len(b) > len(p) {
		b = b[len(b)-PublicKeyLength:]
	}

	copy(p[PublicKeyLength-len(b):], b)
}

//func HexToIdentity(identityHex string) (Identity, error) {
//	if strings.Contains(identityHex, "0x") {
//		identityHex = identityHex[2:]
//	}
//	idBytes, err := hex.DecodeString(identityHex)
//	if err != nil {
//		return [32]byte{}, err
//	}
//	if len(idBytes) != IdentityLength {
//		return [32]byte{}, errors.New("Identity length is not 32 ")
//	}
//
//	var id Identity
//	copy(id[:], idBytes)
//	return id, nil
//}
//
//func HexToHash(hashHex string) (Hash, error) {
//	if strings.Contains(hashHex, "0x") {
//		hashHex = hashHex[2:]
//	}
//	hashBytes, err := hex.DecodeString(hashHex)
//	if err != nil {
//		return [32]byte{}, err
//	}
//	if len(hashBytes) != HashLength {
//		return [32]byte{}, errors.New("Hash length is not 32 ")
//	}
//
//	var hash Hash
//	copy(hash[:], hashBytes)
//	return hash, nil
//}
