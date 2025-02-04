package LocalStructs

type KeepAlive struct {
	ClientIP         string `json:"client_ip" binding:"required"`
	CliendID         string `json:"cliend_id" binding:"required"`
	Timestamp        int64  `json:"timestamp" binding:"required"`
	RequestSignature string `json:"request_signature" binding:"required"`
}

type DataLogin struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	UUID      string `json:"uuid" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required"`
	Sequence  int64  `json:"sequence" binding:"required"`
}
