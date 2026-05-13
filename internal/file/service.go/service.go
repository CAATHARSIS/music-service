package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	commonpb "github.com/CAATHARSIS/music-service/api/gen/common"
	filepb "github.com/CAATHARSIS/music-service/api/gen/file"
	"github.com/CAATHARSIS/music-service/internal/file/models"
	"github.com/CAATHARSIS/music-service/internal/file/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileService struct {
	filepb.UnimplementedFileServiceServer
	pgRepo    repository.PostgresRepository
	minioRepo repository.MinioRepository
	log       *slog.Logger
}

func NewFileService(pgRepo repository.PostgresRepository, minioRepo repository.MinioRepository, log *slog.Logger) *FileService {
	return &FileService{
		pgRepo:    pgRepo,
		minioRepo: minioRepo,
		log:       log,
	}
}

func (s *FileService) UploadFile(stream filepb.FileService_UploadFileServer) error {
	var metadata *filepb.FileMetadata
	var buf bytes.Buffer

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, "receive failed")
		}

		switch data := req.Data.(type) {
		case *filepb.UploadFileRequest_Metadata:
			metadata = data.Metadata
		case *filepb.UploadFileRequest_Chunk:
			buf.Write(data.Chunk)
		}
	}

	if metadata == nil {
		return status.Error(codes.InvalidArgument, "metadata requried")
	}

	key := fmt.Sprintf("%s/%s", uuid.New().String(), metadata.OriginalName)
	bucket := metadata.Bucket
	if bucket == "" {
		bucket = "music-files"
	}

	err := s.minioRepo.Upload(stream.Context(), key, &buf, int64(buf.Len()), metadata.MimeType)
	if err != nil {
		s.log.Error("upload to minio failed", "error", err)
		return status.Error(codes.Internal, "upload failed")
	}

	file := &models.File{
		ID:           uuid.New().String(),
		OriginalName: metadata.OriginalName,
		Bucket:       bucket,
		Key:          key,
		Size:         int64(buf.Len()),
		MimeType:     metadata.MimeType,
		UploadedBy:   &metadata.UploadedBy,
	}

	if err := s.pgRepo.CreateFile(stream.Context(), file); err != nil {
		return status.Error(codes.Internal, "save metadata failed")
	}

	return stream.SendAndClose(convertFileToProto(file))
}

func (s *FileService) GetFileInfo(ctx context.Context, req *filepb.GetFileInfoRequest) (*filepb.FileInfo, error) {
	file, err := s.pgRepo.GetFile(ctx, req.FileId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}
	return convertFileToProto(file), nil
}

func (s *FileService) GetDownloadURL(ctx context.Context, req *filepb.GetDownloadURLRequest) (*filepb.DownloadURLResponse, error) {
	file, err := s.pgRepo.GetFile(ctx, req.FileId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	expiry := time.Duration(req.ExpirySeconds) * time.Second
	if expiry <= 0 {
		expiry = time.Hour
	}

	url, err := s.minioRepo.GetPresignedURL(ctx, file.Key, expiry)
	if err != nil {
		return nil, status.Error(codes.Internal, "generate url failed")
	}

	return &filepb.DownloadURLResponse{
		Url:       url,
		ExpiresAt: time.Now().Add(expiry).Unix(),
	}, nil
}

func (s *FileService) DeleteFile(ctx context.Context, req *filepb.DeleteFileRequest) (*commonpb.Empty, error) {
	file, err := s.pgRepo.GetFile(ctx, req.FileId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	if err := s.minioRepo.Delete(ctx, file.Key); err != nil {
		return nil, status.Error(codes.Internal, "delete from storage failed")
	}

	s.pgRepo.DeleteFile(ctx, req.FileId)
	return &commonpb.Empty{}, nil
}

func (s *FileService) Health(ctx context.Context, req *commonpb.Empty) (*commonpb.HealthyCheckResponse, error) {
	return &commonpb.HealthyCheckResponse{Status: "SERVING"}, nil
}