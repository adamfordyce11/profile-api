package profile

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

type LocalImageStore struct {
    BasePath string
}

func (l *LocalImageStore) SaveImage(userID, filename string, file multipart.File) (string, error) {
    imageName := fmt.Sprintf("%s-%s", userID, filename)
    imagePath := filepath.Join(l.BasePath, imageName)
    out, err := os.Create(imagePath)
    if err != nil {
        return "", err
    }
    defer out.Close()
    _, err = io.Copy(out, file)
    if err != nil {
        return "", err
    }
    return "/images/" + imageName, nil
}