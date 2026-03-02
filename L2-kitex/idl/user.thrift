namespace go user

struct RegisterReq {
    1: string username
    2: string password
}

struct RegisterResp {
    1: bool success
    2: string message
}

service UserService {
    RegisterResp Register(1: RegisterReq req)
}