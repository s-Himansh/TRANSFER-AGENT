package models

type TransferMetaData struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	CheckSum string `json:"check_sum"`
}
