package auth

import (
	"errors"
	"strings"

	"github.com/shubhvish4495/basilisk/pkg/user"
)

var (
	ErrMalformedRole  = errors.New("malformed roles")
	ErrNoRoleSupplied = errors.New("no role attached with user")
	ErrNoRoleFound    = errors.New("user has no role")
	ErrRoleArrayEmpty = errors.New("roles to check array is empty")
)

// getRoleOperation maps a string representation of an operation to a corresponding
// user.RoleOperation constant. It supports the following mappings:
// - "read" maps to user.Read
// - "create" maps to user.Create
// - "*" maps to user.Admin
// If the input string does not match any of the predefined cases, it defaults to user.NoOp.
//
// Parameters:
//   - op: A string representing the operation.
//
// Returns:
//   - A user.RoleOperation value corresponding to the input string.
func getRoleOperation(op string) user.RoleOperation {
	switch op {
	case "read":
		return user.Read
	case "create":
		return user.Create
	case "*":
		return user.Admin
	}
	return user.NoOp
}

// GetRoleString converts a slice of user.Role objects into a slice of strings,
// where each string represents a role in the format "Service:Resource:Operation".
// The Operation is represented as "read", "create", "*", or an empty string
// depending on the value of the user.Role.Operation field.
//
// Parameters:
//   - roles: A slice of user.Role objects to be converted.
//
// Returns:
//   - A slice of strings where each string represents a role in the specified format.
func GetRoleString(roles []user.Role) []string {
	strRoles := make([]string, 0)
	for _, r := range roles {
		opStr := ""
		roleStr := r.Service
		roleStr += ":" + r.Resource

		switch r.Operation {
		case user.Read:
			opStr = "read"
		case user.Create:
			opStr = "create"
		case user.Admin:
			opStr = "*"
		case user.NoOp:
			opStr = ""
		}

		roleStr += ":" + opStr
		strRoles = append(strRoles, roleStr)
	}
	return strRoles
}

// ExtractRoles parses a slice of role strings into a slice of user.Role objects.
// Each role string is expected to be in the format "Service:Resource:Operation".
//
// Parameters:
//   - rS: A slice of strings representing roles.
//
// Returns:
//   - A slice of user.Role objects if parsing is successful.
//   - An error if any role string is malformed or if no roles are supplied.
//
// Errors:
//   - ErrMalformedRole: Returned if a role string does not have at least three parts.
//   - ErrNoRoleSupplied: Returned if the input slice is empty or no valid roles are parsed.
func ExtractRoles(rS []string) ([]user.Role, error) {
	rolesArr := make([]user.Role, 0)
	for _, s := range rS {
		rArr := strings.Split(s, ":")
		if len(rArr) < 3 {
			return rolesArr, ErrMalformedRole
		}

		rolesArr = append(rolesArr, user.Role{
			Service:   rArr[0],
			Resource:  rArr[1],
			Operation: getRoleOperation(rArr[2]),
		})
	}
	if len(rolesArr) == 0 {
		return rolesArr, ErrNoRoleSupplied
	}
	return rolesArr, nil
}

// CheckAccessBasedOnRole determines if a given role has access based on a list of roles.
// It evaluates whether the specified role (`roleToCheck`) is allowed access by checking
// against the provided roles (`rolesPassed`). The function considers the following conditions:
//   - The service of the role matches or is a wildcard ("*").
//   - The resource of the role matches or is a wildcard ("*").
//   - The operation level of the role is greater than or equal to the required operation level.
//
// Parameters:
//   - rolesPassed: A slice of user.Role representing the roles to check against.
//   - roleToCheck: A user.Role representing the role to validate access for.
//
// Returns:
//   - A boolean indicating whether access is granted.
//   - An error if the rolesPassed slice is empty.
//
// Errors:
//   - ErrRoleArrayEmpty: Returned when the rolesPassed slice is empty.
func CheckAccessBasedOnRole(rolesPassed []user.Role, roleToCheck user.Role) (bool, error) {
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
