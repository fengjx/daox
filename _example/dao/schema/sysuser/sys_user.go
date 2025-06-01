package sysuser

import (
	"github.com/fengjx/daox/v2/sqlbuilder"
	"github.com/fengjx/daox/v2/sqlbuilder/ql"

	"time"
)

const (
	TableName     = "sys_user"
	IDField       = "id"
	UsernameField = "username"
	PwdField      = "pwd"
	SaltField     = "salt"
	EmailField    = "email"
	NicknameField = "nickname"
	AvatarField   = "avatar"
	PhoneField    = "phone"
	StatusField   = "status"
	RemarkField   = "remark"
	UtimeField    = "utime"
	CtimeField    = "ctime"
)

var Meta = sysUserMeta{}

// SysUserM 用户信息表
type sysUserMeta struct {
}

func (m sysUserMeta) TableName() string {
	return TableName
}

func (m sysUserMeta) IsAutoIncrement() bool {
	return true
}

func (m sysUserMeta) PrimaryKey() string {
	return "id"
}

func (m sysUserMeta) Columns() []string {
	return []string{
		"id",
		"username",
		"pwd",
		"salt",
		"email",
		"nickname",
		"avatar",
		"phone",
		"status",
		"remark",
		"utime",
		"ctime",
	}
}

func IdIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(IDField).In(args...)
}

func IdNotIn(vals ...int64) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(IDField).NotIn(args...)
}

func IdEQ(val int64) sqlbuilder.Column {
	return ql.Col(IDField).EQ(val)
}

func IdNotEQ(val int64) sqlbuilder.Column {
	return ql.Col(IDField).NotEQ(val)
}

func IdLT(val int64) sqlbuilder.Column {
	return ql.Col(IDField).LT(val)
}

func IdLTEQ(val int64) sqlbuilder.Column {
	return ql.Col(IDField).LTEQ(val)
}

func IdGT(val int64) sqlbuilder.Column {
	return ql.Col(IDField).GT(val)
}

func IdGTEQ(val int64) sqlbuilder.Column {
	return ql.Col(IDField).GTEQ(val)
}

func IdLike(val int64) sqlbuilder.Column {
	return ql.Col(IDField).Like(val)
}

func IdNotLike(val int64) sqlbuilder.Column {
	return ql.Col(IDField).NotLike(val)
}

func IdDesc() sqlbuilder.OrderBy {
	return ql.Desc(IDField)
}

func IdAsc() sqlbuilder.OrderBy {
	return ql.Asc(IDField)
}

func UsernameIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(UsernameField).In(args...)
}

func UsernameNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(UsernameField).NotIn(args...)
}

func UsernameEQ(val string) sqlbuilder.Column {
	return ql.Col(UsernameField).EQ(val)
}

func UsernameNotEQ(val string) sqlbuilder.Column {
	return ql.Col(UsernameField).NotEQ(val)
}

func UsernameLT(val string) sqlbuilder.Column {
	return ql.Col(UsernameField).LT(val)
}

func UsernameLTEQ(val string) sqlbuilder.Column {
	return ql.Col(UsernameField).LTEQ(val)
}

func UsernameGT(val string) sqlbuilder.Column {
	return ql.Col(UsernameField).GT(val)
}

func UsernameGTEQ(val string) sqlbuilder.Column {
	return ql.Col(UsernameField).GTEQ(val)
}

func UsernameLike(val string) sqlbuilder.Column {
	return ql.Col(UsernameField).Like(val)
}

func UsernameNotLike(val string) sqlbuilder.Column {
	return ql.Col(UsernameField).NotLike(val)
}

func UsernameDesc() sqlbuilder.OrderBy {
	return ql.Desc(UsernameField)
}

func UsernameAsc() sqlbuilder.OrderBy {
	return ql.Asc(UsernameField)
}

func PwdIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(PwdField).In(args...)
}

func PwdNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(PwdField).NotIn(args...)
}

func PwdEQ(val string) sqlbuilder.Column {
	return ql.Col(PwdField).EQ(val)
}

func PwdNotEQ(val string) sqlbuilder.Column {
	return ql.Col(PwdField).NotEQ(val)
}

func PwdLT(val string) sqlbuilder.Column {
	return ql.Col(PwdField).LT(val)
}

func PwdLTEQ(val string) sqlbuilder.Column {
	return ql.Col(PwdField).LTEQ(val)
}

func PwdGT(val string) sqlbuilder.Column {
	return ql.Col(PwdField).GT(val)
}

func PwdGTEQ(val string) sqlbuilder.Column {
	return ql.Col(PwdField).GTEQ(val)
}

func PwdLike(val string) sqlbuilder.Column {
	return ql.Col(PwdField).Like(val)
}

func PwdNotLike(val string) sqlbuilder.Column {
	return ql.Col(PwdField).NotLike(val)
}

func PwdDesc() sqlbuilder.OrderBy {
	return ql.Desc(PwdField)
}

func PwdAsc() sqlbuilder.OrderBy {
	return ql.Asc(PwdField)
}

func SaltIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(SaltField).In(args...)
}

func SaltNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(SaltField).NotIn(args...)
}

func SaltEQ(val string) sqlbuilder.Column {
	return ql.Col(SaltField).EQ(val)
}

func SaltNotEQ(val string) sqlbuilder.Column {
	return ql.Col(SaltField).NotEQ(val)
}

func SaltLT(val string) sqlbuilder.Column {
	return ql.Col(SaltField).LT(val)
}

func SaltLTEQ(val string) sqlbuilder.Column {
	return ql.Col(SaltField).LTEQ(val)
}

func SaltGT(val string) sqlbuilder.Column {
	return ql.Col(SaltField).GT(val)
}

func SaltGTEQ(val string) sqlbuilder.Column {
	return ql.Col(SaltField).GTEQ(val)
}

func SaltLike(val string) sqlbuilder.Column {
	return ql.Col(SaltField).Like(val)
}

func SaltNotLike(val string) sqlbuilder.Column {
	return ql.Col(SaltField).NotLike(val)
}

func SaltDesc() sqlbuilder.OrderBy {
	return ql.Desc(SaltField)
}

func SaltAsc() sqlbuilder.OrderBy {
	return ql.Asc(SaltField)
}

func EmailIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(EmailField).In(args...)
}

func EmailNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(EmailField).NotIn(args...)
}

func EmailEQ(val string) sqlbuilder.Column {
	return ql.Col(EmailField).EQ(val)
}

func EmailNotEQ(val string) sqlbuilder.Column {
	return ql.Col(EmailField).NotEQ(val)
}

func EmailLT(val string) sqlbuilder.Column {
	return ql.Col(EmailField).LT(val)
}

func EmailLTEQ(val string) sqlbuilder.Column {
	return ql.Col(EmailField).LTEQ(val)
}

func EmailGT(val string) sqlbuilder.Column {
	return ql.Col(EmailField).GT(val)
}

func EmailGTEQ(val string) sqlbuilder.Column {
	return ql.Col(EmailField).GTEQ(val)
}

func EmailLike(val string) sqlbuilder.Column {
	return ql.Col(EmailField).Like(val)
}

func EmailNotLike(val string) sqlbuilder.Column {
	return ql.Col(EmailField).NotLike(val)
}

func EmailDesc() sqlbuilder.OrderBy {
	return ql.Desc(EmailField)
}

func EmailAsc() sqlbuilder.OrderBy {
	return ql.Asc(EmailField)
}

func NicknameIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(NicknameField).In(args...)
}

func NicknameNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(NicknameField).NotIn(args...)
}

func NicknameEQ(val string) sqlbuilder.Column {
	return ql.Col(NicknameField).EQ(val)
}

func NicknameNotEQ(val string) sqlbuilder.Column {
	return ql.Col(NicknameField).NotEQ(val)
}

func NicknameLT(val string) sqlbuilder.Column {
	return ql.Col(NicknameField).LT(val)
}

func NicknameLTEQ(val string) sqlbuilder.Column {
	return ql.Col(NicknameField).LTEQ(val)
}

func NicknameGT(val string) sqlbuilder.Column {
	return ql.Col(NicknameField).GT(val)
}

func NicknameGTEQ(val string) sqlbuilder.Column {
	return ql.Col(NicknameField).GTEQ(val)
}

func NicknameLike(val string) sqlbuilder.Column {
	return ql.Col(NicknameField).Like(val)
}

func NicknameNotLike(val string) sqlbuilder.Column {
	return ql.Col(NicknameField).NotLike(val)
}

func NicknameDesc() sqlbuilder.OrderBy {
	return ql.Desc(NicknameField)
}

func NicknameAsc() sqlbuilder.OrderBy {
	return ql.Asc(NicknameField)
}

func AvatarIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(AvatarField).In(args...)
}

func AvatarNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(AvatarField).NotIn(args...)
}

func AvatarEQ(val string) sqlbuilder.Column {
	return ql.Col(AvatarField).EQ(val)
}

func AvatarNotEQ(val string) sqlbuilder.Column {
	return ql.Col(AvatarField).NotEQ(val)
}

func AvatarLT(val string) sqlbuilder.Column {
	return ql.Col(AvatarField).LT(val)
}

func AvatarLTEQ(val string) sqlbuilder.Column {
	return ql.Col(AvatarField).LTEQ(val)
}

func AvatarGT(val string) sqlbuilder.Column {
	return ql.Col(AvatarField).GT(val)
}

func AvatarGTEQ(val string) sqlbuilder.Column {
	return ql.Col(AvatarField).GTEQ(val)
}

func AvatarLike(val string) sqlbuilder.Column {
	return ql.Col(AvatarField).Like(val)
}

func AvatarNotLike(val string) sqlbuilder.Column {
	return ql.Col(AvatarField).NotLike(val)
}

func AvatarDesc() sqlbuilder.OrderBy {
	return ql.Desc(AvatarField)
}

func AvatarAsc() sqlbuilder.OrderBy {
	return ql.Asc(AvatarField)
}

func PhoneIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(PhoneField).In(args...)
}

func PhoneNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(PhoneField).NotIn(args...)
}

func PhoneEQ(val string) sqlbuilder.Column {
	return ql.Col(PhoneField).EQ(val)
}

func PhoneNotEQ(val string) sqlbuilder.Column {
	return ql.Col(PhoneField).NotEQ(val)
}

func PhoneLT(val string) sqlbuilder.Column {
	return ql.Col(PhoneField).LT(val)
}

func PhoneLTEQ(val string) sqlbuilder.Column {
	return ql.Col(PhoneField).LTEQ(val)
}

func PhoneGT(val string) sqlbuilder.Column {
	return ql.Col(PhoneField).GT(val)
}

func PhoneGTEQ(val string) sqlbuilder.Column {
	return ql.Col(PhoneField).GTEQ(val)
}

func PhoneLike(val string) sqlbuilder.Column {
	return ql.Col(PhoneField).Like(val)
}

func PhoneNotLike(val string) sqlbuilder.Column {
	return ql.Col(PhoneField).NotLike(val)
}

func PhoneDesc() sqlbuilder.OrderBy {
	return ql.Desc(PhoneField)
}

func PhoneAsc() sqlbuilder.OrderBy {
	return ql.Asc(PhoneField)
}

func StatusIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(StatusField).In(args...)
}

func StatusNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(StatusField).NotIn(args...)
}

func StatusEQ(val string) sqlbuilder.Column {
	return ql.Col(StatusField).EQ(val)
}

func StatusNotEQ(val string) sqlbuilder.Column {
	return ql.Col(StatusField).NotEQ(val)
}

func StatusLT(val string) sqlbuilder.Column {
	return ql.Col(StatusField).LT(val)
}

func StatusLTEQ(val string) sqlbuilder.Column {
	return ql.Col(StatusField).LTEQ(val)
}

func StatusGT(val string) sqlbuilder.Column {
	return ql.Col(StatusField).GT(val)
}

func StatusGTEQ(val string) sqlbuilder.Column {
	return ql.Col(StatusField).GTEQ(val)
}

func StatusLike(val string) sqlbuilder.Column {
	return ql.Col(StatusField).Like(val)
}

func StatusNotLike(val string) sqlbuilder.Column {
	return ql.Col(StatusField).NotLike(val)
}

func StatusDesc() sqlbuilder.OrderBy {
	return ql.Desc(StatusField)
}

func StatusAsc() sqlbuilder.OrderBy {
	return ql.Asc(StatusField)
}

func RemarkIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(RemarkField).In(args...)
}

func RemarkNotIn(vals ...string) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(RemarkField).NotIn(args...)
}

func RemarkEQ(val string) sqlbuilder.Column {
	return ql.Col(RemarkField).EQ(val)
}

func RemarkNotEQ(val string) sqlbuilder.Column {
	return ql.Col(RemarkField).NotEQ(val)
}

func RemarkLT(val string) sqlbuilder.Column {
	return ql.Col(RemarkField).LT(val)
}

func RemarkLTEQ(val string) sqlbuilder.Column {
	return ql.Col(RemarkField).LTEQ(val)
}

func RemarkGT(val string) sqlbuilder.Column {
	return ql.Col(RemarkField).GT(val)
}

func RemarkGTEQ(val string) sqlbuilder.Column {
	return ql.Col(RemarkField).GTEQ(val)
}

func RemarkLike(val string) sqlbuilder.Column {
	return ql.Col(RemarkField).Like(val)
}

func RemarkNotLike(val string) sqlbuilder.Column {
	return ql.Col(RemarkField).NotLike(val)
}

func RemarkDesc() sqlbuilder.OrderBy {
	return ql.Desc(RemarkField)
}

func RemarkAsc() sqlbuilder.OrderBy {
	return ql.Asc(RemarkField)
}

func UtimeIn(vals ...time.Time) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(UtimeField).In(args...)
}

func UtimeNotIn(vals ...time.Time) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(UtimeField).NotIn(args...)
}

func UtimeEQ(val time.Time) sqlbuilder.Column {
	return ql.Col(UtimeField).EQ(val)
}

func UtimeNotEQ(val time.Time) sqlbuilder.Column {
	return ql.Col(UtimeField).NotEQ(val)
}

func UtimeLT(val time.Time) sqlbuilder.Column {
	return ql.Col(UtimeField).LT(val)
}

func UtimeLTEQ(val time.Time) sqlbuilder.Column {
	return ql.Col(UtimeField).LTEQ(val)
}

func UtimeGT(val time.Time) sqlbuilder.Column {
	return ql.Col(UtimeField).GT(val)
}

func UtimeGTEQ(val time.Time) sqlbuilder.Column {
	return ql.Col(UtimeField).GTEQ(val)
}

func UtimeLike(val time.Time) sqlbuilder.Column {
	return ql.Col(UtimeField).Like(val)
}

func UtimeNotLike(val time.Time) sqlbuilder.Column {
	return ql.Col(UtimeField).NotLike(val)
}

func UtimeDesc() sqlbuilder.OrderBy {
	return ql.Desc(UtimeField)
}

func UtimeAsc() sqlbuilder.OrderBy {
	return ql.Asc(UtimeField)
}

func CtimeIn(vals ...time.Time) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(CtimeField).In(args...)
}

func CtimeNotIn(vals ...time.Time) sqlbuilder.Column {
	var args []any
	for _, val := range vals {
		args = append(args, val)
	}
	return ql.Col(CtimeField).NotIn(args...)
}

func CtimeEQ(val time.Time) sqlbuilder.Column {
	return ql.Col(CtimeField).EQ(val)
}

func CtimeNotEQ(val time.Time) sqlbuilder.Column {
	return ql.Col(CtimeField).NotEQ(val)
}

func CtimeLT(val time.Time) sqlbuilder.Column {
	return ql.Col(CtimeField).LT(val)
}

func CtimeLTEQ(val time.Time) sqlbuilder.Column {
	return ql.Col(CtimeField).LTEQ(val)
}

func CtimeGT(val time.Time) sqlbuilder.Column {
	return ql.Col(CtimeField).GT(val)
}

func CtimeGTEQ(val time.Time) sqlbuilder.Column {
	return ql.Col(CtimeField).GTEQ(val)
}

func CtimeLike(val time.Time) sqlbuilder.Column {
	return ql.Col(CtimeField).Like(val)
}

func CtimeNotLike(val time.Time) sqlbuilder.Column {
	return ql.Col(CtimeField).NotLike(val)
}

func CtimeDesc() sqlbuilder.OrderBy {
	return ql.Desc(CtimeField)
}

func CtimeAsc() sqlbuilder.OrderBy {
	return ql.Asc(CtimeField)
}
