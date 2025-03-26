package auth

import (
	"reflect"
	"testing"
)

func TestCheckAccessBasedOnRole(t *testing.T) {
	type args struct {
		rolesPassed []Role
		roleToCheck Role
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
				rolesPassed: []Role{
					{Service: "*", Resource: "*", Operation: admin},
				},
				roleToCheck: Role{
					Service:   "user",
					Resource:  "users",
					Operation: create,
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Check role denial - operation lower role",
			args: args{
				rolesPassed: []Role{
					{Service: "user", Resource: "users", Operation: read},
				},
				roleToCheck: Role{
					Service:   "user",
					Resource:  "users",
					Operation: create,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Check role denial - resource mismatch",
			args: args{
				rolesPassed: []Role{
					{Service: "user", Resource: "customers", Operation: read},
				},
				roleToCheck: Role{
					Service:   "user",
					Resource:  "users",
					Operation: create,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Check role denial - Service mismatch",
			args: args{
				rolesPassed: []Role{
					{Service: "customer", Resource: "users", Operation: read},
				},
				roleToCheck: Role{
					Service:   "user",
					Resource:  "users",
					Operation: create,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "No Role Passed",
			args: args{
				rolesPassed: []Role{},
				roleToCheck: Role{
					Service:   "user",
					Resource:  "users",
					Operation: create,
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
		want    []Role
		wantErr bool
	}{
		{
			name: "Role extraction success - create",
			args: args{
				rS: []string{"user-service:users:create"},
			},
			want: []Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: create,
				},
			},
			wantErr: false,
		},
		{
			name: "Role extraction success - read",
			args: args{
				rS: []string{"user-service:users:read"},
			},
			want: []Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: read,
				},
			},
			wantErr: false,
		},
		{
			name: "Role extraction success - admin",
			args: args{
				rS: []string{"user-service:users:*"},
			},
			want: []Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: admin,
				},
			},
			wantErr: false,
		},
		{
			name: "Role extraction success - noOp",
			args: args{
				rS: []string{"user-service:users:ad-x"},
			},
			want: []Role{
				{
					Service:   "user-service",
					Resource:  "users",
					Operation: noOp,
				},
			},
			wantErr: false,
		},
		{
			name: "No Role passed",
			args: args{
				rS: []string{},
			},
			want:    []Role{},
			wantErr: true,
		},
		{
			name: "Improper role passed",
			args: args{
				rS: []string{"user-service-test-user"},
			},
			want:    []Role{},
			wantErr: true,
		},
		{
			name: "Extract Admin role",
			args: args{
				rS: []string{"*:*:*"},
			},
			want: []Role{
				{
					Service:   "*",
					Resource:  "*",
					Operation: admin,
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
