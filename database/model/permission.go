package model

const (
	Owner Permission = 1 << iota
	Administrator
	Trustor
)

// 权限
type Permission uint64

func (p Permission) String() string {
	switch p {
	case Owner:
		return "Owner"
	case Administrator:
		return "Administrator"
	case Trustor:
		return "Trustor"
	default:
		return "User"
	}
}

func (p Permission) Is(permissions ...Permission) bool {
	for _, v := range permissions {
		if p&v == 0 {
			return false
		}
	}
	return true
}

func (p Permission) Has(permissions ...Permission) bool {
	for _, v := range permissions {
		if p&v != 0 {
			return true
		}
	}
	return false
}

func (p Permission) IsTrusted() bool {
	return p.Has(Owner, Administrator, Trustor)
}

func (p Permission) IsAdmin() bool {
	return p.Has(Owner, Administrator)
}

func (p Permission) IsOwner() bool {
	return p.Is(Owner)
}

// var ErrHasOwner = errors.New("model: there can only be one owner")
// var ErrAppoint1 = errors.New("model: only the owner can appoint the administrator")
// var ErrAppoint2 = errors.New("model: only the administrator can appoint others")

// func (p Permission) Check(n Permission) error {
// 	if n.Has(Owner) {
// 		return ErrHasOwner
// 	}
// 	if n.Has(Administrator) && !p.Has(Owner) {
// 		return ErrAppoint1
// 	}
// 	if !p.IsAdmin() {
// 		return ErrAppoint2
// 	}
// 	return nil
// }
