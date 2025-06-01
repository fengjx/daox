package kit

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

func TestKebabCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "单个小写单词",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "单个大写单词",
			input:    "HELLO",
			expected: "hello",
		},
		{
			name:     "驼峰命名",
			input:    "helloWorld",
			expected: "hello-world",
		},
		{
			name:     "大写驼峰命名",
			input:    "HelloWorld",
			expected: "hello-world",
		},
		{
			name:     "多个单词",
			input:    "ThisIsALongString",
			expected: "this-is-a-long-string",
		},
		{
			name:     "包含数字",
			input:    "Hello123World",
			expected: "hello123-world",
		},
		{
			name:     "连续大写字母",
			input:    "MyXMLParser",
			expected: "my-xml-parser",
		},
		{
			name:     "已经是 kebab-case",
			input:    "hello-world",
			expected: "hello-world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KebabCase(tt.input)
			if got != tt.expected {
				t.Errorf("KebabCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToLowerAndTrim(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "小写字母",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "大写字母",
			input:    "HELLO",
			expected: "hello",
		},
		{
			name:     "包含连字符",
			input:    "hello-world",
			expected: "helloworld",
		},
		{
			name:     "包含下划线",
			input:    "hello_world",
			expected: "helloworld",
		},
		{
			name:     "包含空格",
			input:    "hello world",
			expected: "helloworld",
		},
		{
			name:     "包含数字",
			input:    "hello123world",
			expected: "hello123world",
		},
		{
			name:     "包含特殊字符",
			input:    "hello@#$%^&*()world",
			expected: "helloworld",
		},
		{
			name:     "混合情况",
			input:    "Hello-World_123@TEST",
			expected: "helloworld123test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToLowerAndTrim(tt.input)
			assert.Equal(t, tt.expected, got, "ToLowerAndTrim(%q) = %q, want %q", tt.input, got, tt.expected)
		})
	}
}
