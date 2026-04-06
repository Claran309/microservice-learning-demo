namespace go post

struct CreatePostReq {
    1: i64 owner
    2: string title
    3: string content
}

struct CreatePostResp {
    1: i64 code
    2: string msg
    3: i64 post_id
    4: bool success
}

struct DeletePostReq {
    1: i64 post_id
}

struct DeletePostResp {
    1: i64 code
    2: string msg
    3: bool success
}

service PostService {
    CreatePostResp CreatePost(1: CreatePostReq req)
    DeletePostResp DeletePost(1: DeletePostReq req)
}
