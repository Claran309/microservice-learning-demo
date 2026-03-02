package internal

type UserRepo struct {
	// 假装有东西
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		// 假装有东西
	}
}

func (repo *UserRepo) AddUser(user User) error {
	return nil
}
