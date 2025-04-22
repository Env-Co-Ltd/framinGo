package framinGo

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/Env-Co-Ltd/framinGo/filesystems"
	"github.com/gabriel-vasile/mimetype"
)

func (f *FraminGo) UploadFile(r *http.Request, destination, fieldName string, fs filesystems.FS) error {
	fileName, err := f.GetFileToUpload(r, fieldName)
	if err != nil {
		f.ErrorLog.Println(err)
		return err
	}

	if fs != nil {
		err = fs.Put(fileName, destination)
		if err != nil {
			f.ErrorLog.Println(err)
			return err
		}
	} else {
		err = os.Rename(fileName, fmt.Sprintf("%s/%s", destination, path.Base(fileName)))
		if err != nil {
			f.ErrorLog.Println(err)
			return err
		}
	}
	defer func() {
		_ = os.Remove(fileName)
	}()

	return nil
}

func (f *FraminGo) GetFileToUpload(r *http.Request, fieldName string) (string, error) {
	_ = r.ParseMultipartForm(f.config.uploads.maxUploadSize)
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	mitType, err := mimetype.DetectReader(file)
	if err != nil {
		f.ErrorLog.Println(err)
		return "", err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		f.ErrorLog.Println(err)
		return "", err
	}

	if !inSlice(f.config.uploads.allowedMineTypes, mitType.String()) {
		return "", errors.New("invalid file type")
	}

	dst, err := os.Create(fmt.Sprintf("./tmp/%s", header.Filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("./tmp/%s", header.Filename), nil
}

func inSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
