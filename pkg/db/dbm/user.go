package dbm

type UserInfo struct {
	Id       uint   `gorm:"column:id;primary_key;not null" json:"id"`         //用户ID
	UserName string `gorm:"column:username;type:varchar(64)" json:"username"` //用户名
	PassWord string `gorm:"column:password;type:varchar(64)" json:"password"` //用户密码
	//Role     string `gorm:"column:role" json:"role"`                  //用户角色
	//Org      string `gorm:"column:org" json:"org"`                    //用户所在的组织机构ID
	//Mobile   uint64 `gorm:"column:mobile" json:"mobile"`              //用户手机号码
	//Email    string `gorm:"column:email" json:"email"`                //用户邮箱
}

func (*UserInfo) TableName() string {
	return "user"
}
