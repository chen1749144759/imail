package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/midoks/imail/internal/tools"
)

type User struct {
	Id       int64  `gorm:"primaryKey"`
	Name     string `gorm:"unique;size:50;comment:登录账户"`
	Nick     string `gorm:"unique;size:50;comment:昵称"`
	Password string `gorm:"size:32;comment:用户密码"`
	Salt     string `gorm:"type:varchar(10)"`
	Code     string `gorm:"size:50;comment:编码"`
	Status   int    `gorm:"comment:状态"`

	IsActive bool
	IsAdmin  bool

	Created     time.Time `gorm:"autoCreateTime;comment:创建时间"`
	CreatedUnix int64     `gorm:"autoCreateTime;comment:创建时间"`
	Updated     time.Time `gorm:"autoCreateTime;comment:更新时间"`
	UpdatedUnix int64     `gorm:"autoCreateTime;comment:更新时间"`
}

func (User) TableName() string {
	return "im_users"
}

func (u *User) GetNick() string {
	return "123123"
}

func (u *User) ValidPassword(oldPwd string) bool {
	inputPwd := tools.Md5(tools.Md5(oldPwd) + u.Salt)

	if strings.EqualFold(u.Password, inputPwd) {
		return true
	}
	return false
}

// RelAvatarLink returns relative avatar link to the site domain,
// which includes app sub-url as prefix. However, it is possible
// to return full URL if user enables Gravatar-like service.
func (User) RelAvatarLink() string {
	return "123123"
	// defaultImgUrl := conf.Server.Subpath + "/img/avatar_default.png"
	// if u.ID == -1 {
	// 	return defaultImgUrl
	// }

	// switch {
	// case u.UseCustomAvatar:
	// 	if !com.IsExist(u.CustomAvatarPath()) {
	// 		return defaultImgUrl
	// 	}
	// 	return fmt.Sprintf("%s/%s/%d", conf.Server.Subpath, USER_AVATAR_URL_PREFIX, u.ID)
	// case conf.Picture.DisableGravatar:
	// 	if !com.IsExist(u.CustomAvatarPath()) {
	// 		if err := u.GenerateRandomAvatar(); err != nil {
	// 			log.Error("GenerateRandomAvatar: %v", err)
	// 		}
	// 	}

	// 	return fmt.Sprintf("%s/%s/%d", conf.Server.Subpath, USER_AVATAR_URL_PREFIX, u.ID)
	// }
	// return tool.AvatarLink(u.AvatarEmail)
}

// CreateUser creates record of a new user.
func CreateUser(u *User) (err error) {
	data := db.First(u, "name = ?", u.Name)
	u.Salt = tools.RandString(10)
	u.Nick = u.Name
	u.Password = tools.Md5(tools.Md5(u.Password) + u.Salt)
	if data.Error != nil {
		result := db.Create(u)
		return result.Error
	}
	return data.Error
}

func UserUpdater(u *User) error {
	r := db.Model(&User{}).Where("id = ?", u.Id).Save(u)
	return r.Error
}

type UserSearchOptions struct {
	Keyword  string
	OrderBy  string
	Page     int
	PageSize int
}

func UserSearchByName(opts *UserSearchOptions) (user []*User, _ int64, _ error) {
	if len(opts.Keyword) == 0 {
		return user, 0, nil
	}

	opts.Keyword = strings.ToLower(opts.Keyword)

	if opts.PageSize <= 0 || opts.PageSize > 20 {
		opts.PageSize = 10
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	searchQuery := "%" + opts.Keyword + "%"
	users := make([]*User, 0, opts.PageSize)

	err := db.Model(&User{}).
		Where("LOWER(name) LIKE ?", searchQuery).
		Or("LOWER(nick) LIKE ?", searchQuery).
		Find(&users)
	return users, UsersCount(), err.Error
}

func UsersList(page, pageSize int) ([]*User, error) {
	users := make([]*User, 0, pageSize)
	err := db.Limit(pageSize).Offset((page - 1) * pageSize).Order("id desc").Find(&users)
	return users, err.Error
}

func UsersCount() int64 {
	var count int64
	db.Model(&User{}).Count(&count)
	return count
}

func UsersVaildCount() int64 {
	var count int64
	db.Model(&User{}).Where("is_active = ?", 1).Count(&count)
	return count
}

func LoginWithCode(name string, code string) (bool, int64) {
	list := strings.SplitN(name, "@", 2)

	var user User
	err := db.First(&user, "name = ?", list[0]).Error

	if err != nil {
		return false, 0
	}

	if user.Code == code {
		return true, user.Id
	}

	return false, 0
}

func LoginByUserPassword(name string, password string) (bool, int64) {
	fmt.Println("name", name)
	var u User
	err := db.First(&u, "name = ?", name).Error

	if err != nil {
		return false, 0
	}

	inputPwd := tools.Md5(tools.Md5(password) + u.Salt)

	fmt.Println(password, inputPwd, u.Password)
	if inputPwd == u.Password {
		return true, u.Id
	}
	return false, 0
}

func UserCheckIsExist(name string) bool {
	var user User
	err := db.First(&user, "name = ?", name).Error
	if err == nil {
		return true
	}
	return false
}

func UserUpdateTokenGetByName(name string, token string) bool {
	db.Model(&User{}).Where("name = ?", name).Update("token", token)
	return true
}

func UserUpdateTokenGetById(id int64, token string) error {
	r := db.Model(&User{}).Where("id = ?", id).Update("token", token)
	return r.Error
}

func UserUpdateCodeGetByName(name string, code string) bool {
	db.Model(&User{}).Where("name = ?", name).Update("code", code)
	return true
}

func UserUpdateCodeGetById(id int64, code string) error {
	r := db.Model(&User{}).Where("id = ?", id).Update("code", code)
	return r.Error
}

func UserUpdateNickGetByName(name string, nick string) error {
	r := db.Model(&User{}).Where("name = ?", name).Update("nick", nick)
	return r.Error
}

func UserGetByName(name string) (User, error) {
	list := strings.SplitN(name, "@", 2)
	var user User
	err := db.First(&user, "name = ?", list[0]).Error
	if err != nil {
		return user, err
	}
	return user, nil
}

func UserGetAdmin() (User, error) {
	var user User
	err := db.Model(&User{}).
		Where("is_active=?", 1).
		Where("is_admin=?", 1).
		Find(&user)
	return user, err.Error
}

func UserGetAdminForName() (string, error) {
	u, err := UserGetAdmin()
	return u.Name, err
}

func UserGetById(id int64) (User, error) {
	var user User
	err := db.First(&user, "id = ?", id).Error
	if err != nil {
		return user, err
	}
	return user, nil
}
