package auth

import (
	"reflect"
	"testing"

	"github.com/shubhvish4495/basilisk/pkg/user"
)

func TestCheckAccessBasedOnRole(t *testing.T) {
	type args struct {
		rolesPassed []user.Role
		roleToCheck user.Role
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Check admin role",
			args: args{
				rolesPassed: []user.Role{
					{Service: "*", Resource: "*", Operation: user.Admin},
				},
				roleToCheck: user.Role{
					Service:   "user",
					Resource:  "users",
					Operation: user.Create,
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Check role denial - operation lower role",
			args: args{
				rolesPassed: []user.Role{
					{Service: "user", Resource: "users", Operation: user.Read},
				},
				roleToCheck: user.Role{
					Service:   "user",
					Resource:  "users",
					Operation: user.Create,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Check role denial - resource mismatch",
			args: args{
				rolesPassed: []user.Role{
					{Service: "user", Resource: "customers", Operation: user.Read},
				},
				roleToCheck: user.Role{
					Service:   "user",
					Resource:  "users",
					Operation: user.Create,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Check role denial - Service mismatch",
			args: args{
				rolesPassed: []user.Role{
					{Service: "customer", Resource: "users", Operation: user.Read},
				},
				roleToCheck: user.Role{
					Service:   "user",
					Resource:  "users",
					Operation: user.Create,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "No Role Passed",
			args: args{
				rolesPassed: []user.Role{},
				roleToCheck: user.Role{
					Service:   "user",
					Resource:  "users",
					Operation: user.Create,
				},
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckAccessBasedOnRole(tt.args.rolesPassed, tt.args.roleToCheck)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAccessBasedOnRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckAccessBasedOnRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractRoles(t *testing.T) {
	type args struct {
		rS []string
	}
	tests := []struct {
		name    string
		args    args
		want    []user.Role
		wantErr bool
	}{
		{
			name: "Role extraction success - create",
			args: args{
				rS: []string{"user-service:users:create"},
			},
			want: []user.Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: user.Create,
				},
			},
			wantErr: false,
		},
		{
			name: "Role extraction success - read",
			args: args{
				rS: []string{"user-service:users:read"},
			},
			want: []user.Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: user.Read,
				},
			},
			wantErr: false,
		},
		{
			name: "Role extraction success - admin",
			args: args{
				rS: []string{"user-service:users:*"},
			},
			want: []user.Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: user.Admin,
				},
			},
			wantErr: false,
		},
		{
			name: "Role extraction success - noOp",
			args: args{
				rS: []string{"user-service:users:ad-x"},
			},
			want: []user.Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: user.NoOp,
				},
			},
			wantErr: false,
		},
		{
			name: "No Role passed",
			args: args{
				rS: []string{},
			},
			want:    []user.Role{},
			wantErr: true,
		},
		{
			name: "Improper role passed",
			args: args{
				rS: []string{"user-service-test-user"},
			},
			want:    []user.Role{},
			wantErr: true,
		},
		{
			name: "Extract Admin role",
			args: args{
				rS: []string{"*:*:*"},
			},
			want: []user.Role{
				{
					Service:   "*",
					Resource:  "*",
					Operation: user.Admin,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractRoles(tt.args.rS)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestGetRoleOperation(t *testing.T) {
	tests := []struct {
		name string
		op   string
		want user.RoleOperation
	}{
		{
			name: "Operation read",
			op:   "read",
			want: user.Read,
		},
		{
			name: "Operation create",
			op:   "create",
			want: user.Create,
		},
		{
			name: "Operation admin",
			op:   "*",
			want: user.Admin,
		},
		{
			name: "Operation noOp - invalid operation",
			op:   "invalid",
			want: user.NoOp,
		},
		{
			name: "Operation noOp - empty string",
			op:   "",
			want: user.NoOp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRoleOperation(tt.op); got != tt.want {
				t.Errorf("getRoleOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestGetRoleString(t *testing.T) {
	tests := []struct {
		name  string
		roles []user.Role
		want  []string
	}{
		{
			name: "Single role - create operation",
			roles: []user.Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: user.Create,
				},
			},
			want: []string{"user-service:users:create"},
		},
		{
			name: "Single role - read operation",
			roles: []user.Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: user.Read,
				},
			},
			want: []string{"user-service:users:read"},
		},
		{
			name: "Single role - admin operation",
			roles: []user.Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: user.Admin,
				},
			},
			want: []string{"user-service:users:*"},
		},
		{
			name: "Single role - noOp operation",
			roles: []user.Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: user.NoOp,
				},
			},
			want: []string{"user-service:users:"},
		},
		{
			name: "Multiple roles",
			roles: []user.Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: user.Create,
				},
				{
					Service:   "order-service",
					Resource:  "orders",
					Operation: user.Read,
				},
				{
					Service:   "admin-service",
					Resource:  "*",
					Operation: user.Admin,
				},
			},
			want: []string{
				"user-service:users:create",
				"order-service:orders:read",
				"admin-service:*:*",
			},
		},
		{
			name:  "Empty roles slice",
			roles: []user.Role{},
			want:  []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRoleString(tt.roles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRoleString() = %v, want %v", got, tt.want)
			}
		})
	}
}
