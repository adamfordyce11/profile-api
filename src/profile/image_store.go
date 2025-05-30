package profile

import (
	"mime/multipart"
)

type ImageStore interface {
    SaveImage(userID, filename string, file multipart.File) (string, error)
}