package repositories

import (
	"errors"
	"strconv"
)

type Address struct {
	Type        string `json:"type"`
	Line1       string `json:"street1"`
	Line2       string `json:"street2,omitempty"`
	City        string `json:"city"`
	State       string `json:"state"`
	Postal      string `json:"postal"`
	*Timestamps `json:"timestamps,omitempty"`
}

var cityStateLookup map[int]Address

func makeCityStateLookup() {
	cityStateLookup = make(map[int]Address)
	reno := Address{
		City:  "Reno",
		State: "NV",
	}
	carson := Address{
		City:  "Carson City",
		State: "NV",
	}
	minden := Address{
		City:  "Minden",
		State: "NV",
	}
	cityStateLookup[89511] = reno
	cityStateLookup[89521] = reno
	cityStateLookup[89701] = carson
	cityStateLookup[89706] = carson
	cityStateLookup[89703] = carson
	cityStateLookup[89702] = carson
	cityStateLookup[89423] = minden
}

func (a *Address) FillCityStateFromZip() error {
	postal, err := strconv.Atoi(a.Postal)
	if err != nil {
		return err
	}
	cityState, exists := cityStateLookup[postal]
	if !exists {
		return errors.New("City state for postal code not found!")
	}

	a.City = cityState.City
	a.State = cityState.State

	return nil
}
