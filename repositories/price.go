package repositories

import "fmt"

type Price float32

func (p Price) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("%.2f", float32(p))
	return []byte(s), nil
}
