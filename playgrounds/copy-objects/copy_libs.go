package copyobjects

import (
	"github.com/jinzhu/copier"
	"github.com/mitchellh/copystructure"
)

// CopyCopier creates a deep copy using jinzhu/copier library.
// Popular library with struct tag support for customization.
// go get github.com/jinzhu/copier
func (c *Config) CopyCopier() (*Config, error) {
	if c == nil {
		return nil, nil
	}

	var dst Config
	if err := copier.CopyWithOption(&dst, c, copier.Option{DeepCopy: true}); err != nil {
		return nil, err
	}

	return &dst, nil
}

// CopyCopystructure creates a deep copy using mitchellh/copystructure.
// Hashicorp's deep copy library, battle-tested in production.
// go get github.com/mitchellh/copystructure
func (c *Config) CopyCopystructure() (*Config, error) {
	if c == nil {
		return nil, nil
	}

	copied, err := copystructure.Copy(c)
	if err != nil {
		return nil, err
	}

	return copied.(*Config), nil
}
