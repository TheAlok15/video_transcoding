package service

import (
	"context"
	"fmt"
	"mime/multipart"
)


type UploadService struct {

}

func NewUploadService() *UploadService{
	return &UploadService{}
}

func (s *UploadService) UploadVideo(ctx context.Context, fileHeader *multipart.FileHeader, jobID string) error {

	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()


	


}