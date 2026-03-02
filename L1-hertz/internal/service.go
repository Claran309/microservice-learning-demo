package internal

import (
	"errors"
	util "microservicesDemo/utils"
)

type UserService struct {
	UserRepo UserRepo
}

func NewUserService(userRepo UserRepo) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

func (s *UserService) Register(name string, password string) (*User, error) {
	//密码时候否符合格式
	var flagPassword bool
	for i := 0; i < len(password); i++ {
		if !((password[i] >= 'a' && password[i] <= 'z') || (password[i] >= '0' && password[i] <= '9') || (password[i] >= 'A' && password[i] <= 'Z')) {
			flagPassword = true
		}
	}
	if flagPassword {
		return nil, errors.New("password format Error")
	}

	//加密密码
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		return nil, errors.New("password hash failed" + err.Error())
	}

	//创建用户
	user := User{
		Name:     name,
		Password: hashedPassword, // 加密存储
	}

	//传入数据库
	if err := s.UserRepo.AddUser(user); err != nil {
		return nil, err
	}

	return &user, nil
}
