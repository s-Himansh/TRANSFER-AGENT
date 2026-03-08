package models

import (
	"fmt"
	"time"

	"transfer.agent/utils"
)

type TransferStatus string

type TransferMetaData struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	CheckSum string `json:"check_sum"`
}

type TransferJob struct {
	ID              string
	SourcePath      string // Path to source directory
	DestinationName string // destination file name if needs to be given explicitly
	Priority        int    // to faciltate prioritising transfers
	Status          TransferStatus

	// File Meta data
	TransferMetaData

	// Time related meta
	SubmittedAt *time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time

	//Final state of transfer
	Error            string
	BytesTransferred int64
}

func NewTransferJob(path string) *TransferJob {
	return &TransferJob{
		ID:          generateID(),
		SourcePath:  path,
		Status:      StatusPending,
		SubmittedAt: utils.Address(time.Now()).(*time.Time),
	}
}

func generateID() string {
	return fmt.Sprintf("TRSFR-%d", time.Now().UnixNano())
}
