package crypto

import "golang.org/x/crypto/bcrypt"

// BcryptHasher bcrypt哈希实现
type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) Hasher {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(data string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data), h.cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (h *BcryptHasher) Verify(data, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(data))
	if err != nil {
		return false, nil
	}
	return true, nil
}
