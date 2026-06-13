package permission

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/i18n"
)

func errEmptyXUID() error {
	return fmt.Errorf("%s", i18n.R("%df.permission.error.empty_xuid"))
}

func errEmptyPermission() error {
	return fmt.Errorf("%s", i18n.R("%df.permission.error.empty_permission"))
}
