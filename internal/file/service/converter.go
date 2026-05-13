package service

import (
	"time"

	filepb "github.com/CAATHARSIS/music-service/api/gen/file"
	"github.com/CAATHARSIS/music-service/internal/file/models"
)

func convertFileToProto(file *models.File) *filepb.FileInfo {
	pb := &filepb.FileInfo{
		Id: file.ID,
		OriginalName: file.OriginalName,
		Bucket: file.Bucket,
		Key: file.Key,
		Size: file.Size,
		MimeType: file.MimeType,
		CreatedAt: file.CreatedAt.Format(time.RFC3339),
	}
	
	if file.UploadedBy != nil {
		pb.UploadedBy = *file.UploadedBy
	}

	return pb
}