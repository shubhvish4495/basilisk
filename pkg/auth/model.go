package auth

type RoleOperation int

const (
	NoOp RoleOperation = iota //0
	Read
	Create
	Admin
)

type User struct {
	UUID           string   `json:"uuid"`
	Type           int      `json:"type_id"`
	RolesExtracted []Role   `json:"-"`
	Roles          []string `json:"roles"`
}

type Role struct {
	Service   string
	Resource  string
	Operation RoleOperation
}
