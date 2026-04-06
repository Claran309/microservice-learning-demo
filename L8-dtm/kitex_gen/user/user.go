package user

type RegisterReq struct {
	Username string `thrift:"username,1" json:"username"`
	Password string `thrift:"password,2" json:"password"`
	Email    string `thrift:"email,3" json:"email"`
}

func NewRegisterReq() *RegisterReq {
	return &RegisterReq{}
}

func (p *RegisterReq) GetUsername() string {
	return p.Username
}

func (p *RegisterReq) GetPassword() string {
	return p.Password
}

func (p *RegisterReq) GetEmail() string {
	return p.Email
}

type RegisterResp struct {
	Code    int64  `thrift:"code,1" json:"code"`
	Msg     string `thrift:"msg,2" json:"msg"`
	UserId  int64  `thrift:"user_id,3" json:"user_id"`
	Success bool   `thrift:"success,4" json:"success"`
}

func NewRegisterResp() *RegisterResp {
	return &RegisterResp{}
}

func (p *RegisterResp) GetCode() int64 {
	return p.Code
}

func (p *RegisterResp) GetMsg() string {
	return p.Msg
}

func (p *RegisterResp) GetUserId() int64 {
	return p.UserId
}

func (p *RegisterResp) GetSuccess() bool {
	return p.Success
}

type LoginReq struct {
	Username string `thrift:"username,1" json:"username"`
	Password string `thrift:"password,2" json:"password"`
}

func NewLoginReq() *LoginReq {
	return &LoginReq{}
}

func (p *LoginReq) GetUsername() string {
	return p.Username
}

func (p *LoginReq) GetPassword() string {
	return p.Password
}

type LoginResp struct {
	Code    int64  `thrift:"code,1" json:"code"`
	Msg     string `thrift:"msg,2" json:"msg"`
	UserId  int64  `thrift:"user_id,3" json:"user_id"`
	Success bool   `thrift:"success,4" json:"success"`
}

func NewLoginResp() *LoginResp {
	return &LoginResp{}
}

func (p *LoginResp) GetCode() int64 {
	return p.Code
}

func (p *LoginResp) GetMsg() string {
	return p.Msg
}

func (p *LoginResp) GetUserId() int64 {
	return p.UserId
}

func (p *LoginResp) GetSuccess() bool {
	return p.Success
}
