package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

func GetUserList(user *models.User, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.User{})

	if user.Username != "" {
		db = db.Where("username = ?", user.Username)
	}
	if user.Nickname != "" {
		db = db.Where("nickname = ?", user.Nickname)
	}

	db = db.Preload("Roles").Omit("password")

	return Pagination(db, []models.User{}, pn, pSize)
}

func GetUserById(id int) (*models.User, error) {
	db := global.Db.Model(&models.User{})

	data := &models.User{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveUser(user *models.User) (*models.User, error) {
	err := IfUserByUserName(user.Username)
	if err != nil {
		return nil, err
	}

	user.Password = utils.GenMd5(user.Password)
	err = global.Db.Model(&models.User{}).Create(user).Error

	return user, err
}

func UpdateUser(user *models.User) (*models.User, error) {
	if user.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetUserById(user.ID)
	if err != nil {
		return nil, err
	}

	user.Username = ""
	user.Password = ""
	user.Roles = nil

	return user, global.Db.Updates(&user).Error
}

func DelUser(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetUserById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	return global.Db.Delete(&data).Error
}

func CheckPassword(username, password string) (*models.User, error) {
	user := &models.User{}

	db := global.Db.Model(&models.User{})
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	if user.Password == utils.GenMd5(password) {
		return user, nil
	}
	return nil, errors.New("wrong password")
}

// ChangePassword 修改密码
func ChangePassword(id int, oldPw, newPw, username string) error {
	user, err := GetUserById(id)
	if err != nil {
		return err
	}

	if user.Password != utils.GenMd5(oldPw) {
		return errors.New("wrong password")
	}
	user.Operator = username
	user.Password = utils.GenMd5(newPw)

	return global.Db.Updates(&user).Error
}

// GetUserFieldList 获取字段列表
func GetUserFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.User{})
	switch field {
	case "username":
		db.Select("username")
	case "nickname":
		db.Select("nickname")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}

// IfUserByUserName 判断用户名是否已存在
func IfUserByUserName(username string) error {
	var count int64
	err := global.Db.Model(&models.User{}).Where("username = ?",
		username).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user name already exists")
	}

	return nil
}

// SetRoles 分配角色
func SetRoles(id int, roleIds []int, operator string) error {
	roles, err := GetRoleByIdList(roleIds)
	if err != nil {
		return err
	}

	user, err := GetUserById(id)
	if err != nil {
		return err
	}

	err = global.Db.Model(&user).Association("Roles").Clear()
	if err != nil {
		return err
	}

	user.Roles = roles
	user.Operator = operator
	return global.Db.Save(&user).Error
}

// GetRolePermissions 获取权限列表
func GetRolePermissions(id int) (interface{}, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}

	db := global.Db.Model(&models.User{})

	data := &models.User{}
	err := db.Where("id = ?", id).Preload("Roles").First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("role does not exist")
	}

	ids := make([]int, 0)
	for _, role := range data.Roles {
		ids = append(ids, role.ID)
	}
	permissions, err := GetPermissions(ids)
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// GetRoles 获取角色列表
func GetRoles(id int) (interface{}, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}

	db := global.Db.Model(&models.User{})

	data := &models.User{}
	err := db.Where("id = ?", id).Preload("Roles").First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("role does not exist")
	}

	ids := make([]int, 0)
	for _, role := range data.Roles {
		ids = append(ids, role.ID)
	}

	return ids, nil
}

func getAdmin(userId int) (bool, error) {
	var total int64
	err := global.Db.Table("tb_user_role").Where(
		"role_id = 1 and user_id = ?", userId).Count(&total).Error
	if err != nil {
		return false, err
	}
	if total > 0 {
		return true, nil
	}

	return false, nil
}
