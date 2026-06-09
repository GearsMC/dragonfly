package permission

import "fmt"

func errEmptyXUID() error {
	return fmt.Errorf("xuid boş olamaz")
}

func errEmptyPermission() error {
	return fmt.Errorf("permission boş olamaz")
}
