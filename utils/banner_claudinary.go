package utils

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadImage(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	cld := InitCloudinary()

	// Read the file content into memory (Cloudinary accepts a reader)
	ctx := context.Background()

	publicID := fileHeader.Filename
	uploadParams := uploader.UploadParams{
		PublicID:       publicID,
		Folder:         "banners",
	}

	result, err := cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", fmt.Errorf("cloudinary upload error: %w", err)
	}

	return result.SecureURL, nil
}
