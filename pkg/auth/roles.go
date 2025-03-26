package auth

import (
	"errors"
	"strings"
)

type Role struct {
	Service   string
	Resource  string
	Operation RoleOperation
}

type RoleOperation int

const (
	noOp RoleOperation = iota //0
	read
	create
	admin
)

var (
	ErrMalformedRole  = errors.New("malformed roles supplied")
	ErrNoRoleSupplied = errors.New("no role found in string")
	ErrNoRoleFound    = errors.New("requested role not found")
	ErrRoleArrayEmpty = errors.New("roles to check array is empty")
)

func getRoleOperation(op string) RoleOperation {
	switch op {
	case "read":
		return read
	case "create":
		return create
	case "*":
		return admin
	}
	return noOp
}

// extract role will extract role from incoming roles in string
func ExtractRoles(rS []string) ([]Role, error) {
	rolesArr := make([]Role, 0)
	for _, s := range rS {
		rArr := strings.Split(s, ":")
		if len(rArr) < 3 {
			return rolesArr, ErrMalformedRole
		}

		rolesArr = append(rolesArr, Role{
			rArr[0],
			rArr[1],
			getRoleOperation(rArr[2]),
		})
	}
	if len(rolesArr) == 0 {
		return rolesArr, ErrNoRoleSupplied
	}
	return rolesArr, nil
}

// CheckAccessBasedOnRole checks if the requested role is present in the role Array
// supplied. Make sure to supply minimum access role required for any check.
// If role is provided with higher access we will automatically grant user
// access.
func CheckAccessBasedOnRole(rolesPassed []Role, roleToCheck Role) (bool, error) {
	if len(rolesPassed) == 0 {
		return false, ErrRoleArrayEmpty
	}

	// check if role service is equal to what we are looking for or is "*" which means all service
	// check if role resource is equal to what we are looking for or is "*" which means all resources
	// check if role operation is equal to what we are looking for or is "*" which means all operations
	// if all these conditions hold to true for any role in the series of roles we will allow access
	for _, r := range rolesPassed {
		if (roleToCheck.Service == r.Service || r.Service == "*") &&
			(roleToCheck.Resource == r.Resource || r.Resource == "*") && (r.Operation >= roleToCheck.Operation) {
			return true, nil
		}
	}
	return false, nil
}
