package user

type RoleOperation int

const (
	NoOp RoleOperation = iota //0
	Read
	Create
	Admin
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Roles    []Role `json:"roles"`
}

type Role struct {
	Service   string
	Resource  string
	Operation RoleOperation
}
