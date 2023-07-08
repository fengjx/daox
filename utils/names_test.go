package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnakeCase(t *testing.T) {
	t.Log(SnakeCase("UserInfo"))
	assert.Equal(t, "user_info", SnakeCase("UserInfo"))
}

func TestTitleCase(t *testing.T) {
	t.Log(TitleCase("user_info"))
	assert.Equal(t, "UserInfo", TitleCase("user_info"))
}

func TestGonicCase(t *testing.T) {
	t.Log(GonicCase("user_id"))
	assert.Equal(t, "UserID", GonicCase("user_id"))
}

func TestFirstUpper(t *testing.T) {
	val := FirstUpper("user")
	t.Log(val)
	assert.Equal(t, "User", val)
}

func TestFirstLower(t *testing.T) {
	val := FirstLower("UserID")
	t.Log(val)
	assert.Equal(t, "userID", val)
}
