package engine

import "testing"

func TestParseSQLType(t *testing.T) {
	type args struct {
		sql string
	}
	tests := []struct {
		name string
		args args
		want SQLType
	}{
		{
			name: "select",
			args: args{
				sql: "SELECT * FROM user WHERE id = ?",
			},
			want: SELECT,
		},
		{
			name: "insert",
			args: args{
				sql: "INSERT INTO user (name, age) VALUES (?, ?)",
			},
			want: INSERT,
		},
		{
			name: "update",
			args: args{
				sql: "UPDATE user SET name = ? WHERE id = ?",
			},
			want: UPDATE,
		},
		{
			name: "delete",
			args: args{
				sql: "DELETE FROM user WHERE id = ?",
			},
			want: DELETE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseSQLType(tt.args.sql); got != tt.want {
				t.Errorf("ParseSQLType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTableName(t *testing.T) {
	type args struct {
		sql string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "select",
			args: args{
				sql: "SELECT * FROM user WHERE id = ?",
			},
			want: "user",
		},
		{
			name: "insert",
			args: args{
				sql: "INSERT INTO user (name, age) VALUES (?, ?)",
			},
			want: "user",
		},
		{
			name: "update",
			args: args{
				sql: "UPDATE user SET name = ? WHERE id = ?",
			},
			want: "user",
		},
		{
			name: "delete",
			args: args{
				sql: "DELETE FROM user WHERE id = ?",
			},
			want: "user",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTableName(tt.args.sql); got != tt.want {
				t.Fatalf("ParseTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}
