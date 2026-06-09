package permission

// Subject, permission hesaplaması yapılan doğrulanmış hesabı temsil eder.
// Kalıcı yetki kararı XUID üzerinden verilir; isim yalnızca okunabilirlik ve log için kullanılır.
type Subject interface {
	PermissionXUID() string
	PermissionName() string
}

// Calculator, permission kararlarını hesaplayan API'dir.
type Calculator interface {
	CalculatePermission(subject Subject, name string) State
}

// Permission, kayıtlı bir permission düğümüdür.
// Children içindeki değer true ise üst permission verildiğinde çocuk da verilir,
// false ise üst permission verildiğinde çocuk açıkça reddedilir.
type Permission struct {
	Name        string
	Description string
	Children    map[string]bool
}

// New, isim ve açıklama ile yeni bir permission oluşturur.
func New(name, description string) Permission {
	return Permission{Name: name, Description: description}
}

// WithChildren, permission alt izinlerini kopyalayarak ekler.
func (p Permission) WithChildren(children map[string]bool) Permission {
	p.Children = make(map[string]bool, len(children))
	for name, value := range children {
		p.Children[name] = value
	}
	return p
}

// NopCalculator, hiçbir permission vermeyen güvenli varsayılan hesaplayıcıdır.
type NopCalculator struct{}

// CalculatePermission her zaman Undefined döndürür.
func (NopCalculator) CalculatePermission(Subject, string) State {
	return Undefined
}
