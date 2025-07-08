package templates

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

type Dep struct {
	Coordinate string
	Version    string
}

func (d Dep) NonSemver() bool {
	_, err := semver.NewVersion(d.Version)
	return err != nil
}

func (d Dep) EqCoord(other Dep) bool {
	return d.Coordinate == other.Coordinate
}

func (d Dep) String() string {
	return fmt.Sprintf("%s:%s", d.Coordinate, d.Version)
}

// TODO handle hashes differently (in another section)
func (d Dep) IsLater(other Dep) (bool, error) {
	v1, err := semver.NewVersion(d.Version)
	if err != nil {
		return false, err
	}
	v2, err := semver.NewVersion(other.Version)
	if err != nil {
		return false, err
	}
	return v1.GreaterThan(v2), nil
}
