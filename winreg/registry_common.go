package winreg

import "errors"

type RegistryType string

const (
	REG_UNKNOWN RegistryType = "REG_UNKNOWN" //nolint:revive,stylecheck
	REG_SZ      RegistryType = "REG_SZ"      //nolint:revive,stylecheck
	REG_BINARY  RegistryType = "REG_BINARY"  //nolint:revive,stylecheck
	REG_DWORD   RegistryType = "REG_DWORD"   //nolint:revive,stylecheck
	REG_QWORD   RegistryType = "REG_QWORD"   //nolint:revive,stylecheck
)

var (
	ErrUnknownRootKey              = errors.New("unknown root key")
	ErrFailedCreatingOrOpeningKey  = errors.New("failed to create/open key")
	ErrFailedUpdatingExistingValue = errors.New("failed updating existing value")
	ErrMissingRootKey              = errors.New("missing root key")
	ErrUnknownValType              = errors.New("unknown val type")
	ErrFailedToConvertVal          = errors.New("failed to convert value to registry type")
	ErrUnknownRegistryType         = errors.New("unknown registry type")
)
