package post

type CreatePostReq struct {
	Owner   int64  `thrift:"owner,1" json:"owner"`
	Title   string `thrift:"title,2" json:"title"`
	Content string `thrift:"content,3" json:"content"`
}

func NewCreatePostReq() *CreatePostReq {
	return &CreatePostReq{}
}

func (p *CreatePostReq) GetOwner() int64 {
	return p.Owner
}

func (p *CreatePostReq) GetTitle() string {
	return p.Title
}

func (p *CreatePostReq) GetContent() string {
	return p.Content
}

type CreatePostResp struct {
	Code    int64  `thrift:"code,1" json:"code"`
	Msg     string `thrift:"msg,2" json:"msg"`
	PostId  int64  `thrift:"post_id,3" json:"post_id"`
	Success bool   `thrift:"success,4" json:"success"`
}

func NewCreatePostResp() *CreatePostResp {
	return &CreatePostResp{}
}

func (p *CreatePostResp) GetCode() int64 {
	return p.Code
}

func (p *CreatePostResp) GetMsg() string {
	return p.Msg
}

func (p *CreatePostResp) GetPostId() int64 {
	return p.PostId
}

func (p *CreatePostResp) GetSuccess() bool {
	return p.Success
}

type DeletePostReq struct {
	PostId int64 `thrift:"post_id,1" json:"post_id"`
}

func NewDeletePostReq() *DeletePostReq {
	return &DeletePostReq{}
}

func (p *DeletePostReq) GetPostId() int64 {
	return p.PostId
}

type DeletePostResp struct {
	Code    int64  `thrift:"code,1" json:"code"`
	Msg     string `thrift:"msg,2" json:"msg"`
	Success bool   `thrift:"success,3" json:"success"`
}

func NewDeletePostResp() *DeletePostResp {
	return &DeletePostResp{}
}

func (p *DeletePostResp) GetCode() int64 {
	return p.Code
}

func (p *DeletePostResp) GetMsg() string {
	return p.Msg
}

func (p *DeletePostResp) GetSuccess() bool {
	return p.Success
}
