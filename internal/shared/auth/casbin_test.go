package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiToSubject(t *testing.T) {
	result := ApiToSubject(123)
	assert.Equal(t, "api_123", result)
}

func TestMenuToSubject(t *testing.T) {
	result := MenuToSubject(456)
	assert.Equal(t, "menu_456", result)
}

func TestButtonToSubject(t *testing.T) {
	result := ButtonToSubject(789)
	assert.Equal(t, "button_789", result)
}

func TestRoleToSubject(t *testing.T) {
	result := RoleToSubject(100)
	assert.Equal(t, "role_100", result)
}

func TestNewCasbinEnforcer(t *testing.T) {
	enforcer, err := NewCasbinEnforcer()
	assert.NoError(t, err)
	assert.NotNil(t, enforcer)
}

func TestAddPolicies(t *testing.T) {
	ctx := context.Background()
	enforcer, _ := NewCasbinEnforcer()

	rules := [][]string{
		{"api_1", "/api/v1/test", "GET"},
		{"api_2", "/api/v1/test2", "POST"},
	}

	err := AddPolicies(ctx, enforcer, rules)
	assert.NoError(t, err)

	ok, _ := enforcer.Enforce("api_1", "/api/v1/test", "GET")
	assert.True(t, ok)

	ok, _ = enforcer.Enforce("api_2", "/api/v1/test2", "POST")
	assert.True(t, ok)
}

func TestAddPolicies_InvalidRule(t *testing.T) {
	ctx := context.Background()
	enforcer, _ := NewCasbinEnforcer()

	rules := [][]string{
		{"api_1", "/api/v1/test"},
	}

	err := AddPolicies(ctx, enforcer, rules)
	assert.Error(t, err)
}

func TestAddPolicies_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	enforcer, _ := NewCasbinEnforcer()
	rules := [][]string{
		{"api_1", "/api/v1/test", "GET"},
	}

	err := AddPolicies(ctx, enforcer, rules)
	assert.Error(t, err)
}

func TestRemovePolicies(t *testing.T) {
	ctx := context.Background()
	enforcer, _ := NewCasbinEnforcer()

	rules := [][]string{
		{"api_1", "/api/v1/test", "GET"},
		{"api_2", "/api/v1/test2", "POST"},
	}
	err := AddPolicies(ctx, enforcer, rules)
	assert.NoError(t, err)

	err = RemovePolicies(ctx, enforcer, rules)
	assert.NoError(t, err)

	ok, _ := enforcer.Enforce("api_1", "/api/v1/test", "GET")
	assert.False(t, ok)
}

func TestRemovePolicies_InvalidRule(t *testing.T) {
	ctx := context.Background()
	enforcer, _ := NewCasbinEnforcer()

	rules := [][]string{
		{"api_1", "/api/v1/test"},
	}

	err := RemovePolicies(ctx, enforcer, rules)
	assert.Error(t, err)
}

func TestRemovePolicies_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	enforcer, _ := NewCasbinEnforcer()
	rules := [][]string{
		{"api_1", "/api/v1/test", "GET"},
	}

	err := RemovePolicies(ctx, enforcer, rules)
	assert.Error(t, err)
}

func TestAddGroupPolicies(t *testing.T) {
	ctx := context.Background()
	enforcer, _ := NewCasbinEnforcer()

	rules := [][]string{
		{"user_1", "role_admin"},
		{"user_2", "role_user"},
	}

	err := AddGroupPolicies(ctx, enforcer, rules)
	assert.NoError(t, err)
}

func TestAddGroupPolicies_InvalidRule(t *testing.T) {
	ctx := context.Background()
	enforcer, _ := NewCasbinEnforcer()

	rules := [][]string{
		{"user_1"},
	}

	err := AddGroupPolicies(ctx, enforcer, rules)
	assert.Error(t, err)
}

func TestAddGroupPolicies_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	enforcer, _ := NewCasbinEnforcer()
	rules := [][]string{
		{"user_1", "role_admin"},
	}

	err := AddGroupPolicies(ctx, enforcer, rules)
	assert.Error(t, err)
}

func TestRemoveFilteredGroupingPolicy(t *testing.T) {
	ctx := context.Background()
	enforcer, _ := NewCasbinEnforcer()

	rules := [][]string{
		{"user_1", "role_admin"},
		{"user_2", "role_user"},
		{"user_1", "role_user"},
	}
	err := AddGroupPolicies(ctx, enforcer, rules)
	assert.NoError(t, err)

	err = RemoveFilteredGroupingPolicy(ctx, enforcer, 0, "user_1")
	assert.NoError(t, err)

	roles, _ := enforcer.GetRolesForUser("user_1")
	assert.Empty(t, roles)
}

func TestRemoveFilteredGroupingPolicy_InvalidParams(t *testing.T) {
	ctx := context.Background()
	enforcer, _ := NewCasbinEnforcer()

	err := RemoveFilteredGroupingPolicy(ctx, enforcer, -1, "user_1")
	assert.Error(t, err)

	err = RemoveFilteredGroupingPolicy(ctx, enforcer, 2, "user_1")
	assert.Error(t, err)

	err = RemoveFilteredGroupingPolicy(ctx, enforcer, 0, "")
	assert.Error(t, err)
}

func TestRemoveFilteredGroupingPolicy_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	enforcer, _ := NewCasbinEnforcer()

	err := RemoveFilteredGroupingPolicy(ctx, enforcer, 0, "user_1")
	assert.Error(t, err)
}

func TestEnforcer_Enforce(t *testing.T) {
	ctx := context.Background()
	enforcer, _ := NewCasbinEnforcer()

	err := AddPolicies(ctx, enforcer, [][]string{
		{"api_1", "/api/v1/login", "POST"},
	})
	assert.NoError(t, err)

	ok, err := enforcer.Enforce("api_1", "/api/v1/login", "POST")
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = enforcer.Enforce("api_1", "/api/v1/login", "GET")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestEnforcerWithRole(t *testing.T) {
	ctx := context.Background()
	enforcer, _ := NewCasbinEnforcer()

	err := AddGroupPolicies(ctx, enforcer, [][]string{
		{"user_1", "role_admin"},
	})
	assert.NoError(t, err)

	err = AddPolicies(ctx, enforcer, [][]string{
		{"role_admin", "/api/v1/admin", "GET"},
	})
	assert.NoError(t, err)

	ok, err := enforcer.Enforce("user_1", "/api/v1/admin", "GET")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestNewCasbinEnforcer_WithInitialPolicy(t *testing.T) {
	enforcer, err := NewCasbinEnforcer()
	assert.NoError(t, err)

	ok, err := enforcer.Enforce("api_0", "/api/v1/login", "POST")
	assert.NoError(t, err)
	assert.True(t, ok)
}
