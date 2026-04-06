namespace go user

struct RegisterReq {
    1: string username
    2: string password
    3: string email
}

struct RegisterResp {
    1: i64 code
    2: string msg
    3: i64 user_id
    4: bool success
}

struct LoginReq {
    1: string username
    2: string password
}

struct LoginResp {
    1: i64 code
    2: string msg
    3: i64 user_id
    4: bool success
}

service UserService {
    RegisterResp Register(1: RegisterReq req)
    LoginResp Login(1: LoginReq req)
}
