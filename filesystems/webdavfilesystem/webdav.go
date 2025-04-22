package webdavfilesystem

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/Env-Co-Ltd/framinGo/filesystems"
	"github.com/studio-b12/gowebdav"
)

type WebDAV struct {
	URL  string
	User string
	Pass string
}

func (w *WebDAV) getClient() *gowebdav.Client {
	c := gowebdav.NewClient(w.URL, w.User, w.Pass)
	return c
}

func (w *WebDAV) Put(fileName, folderName string) error {
	client := w.getClient()

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	err = client.WriteStream(fmt.Sprintf("%s/%s", folderName, path.Base(fileName)), file, 0664)
	if err != nil {
		return err
	}

	return nil
}

func (w *WebDAV) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing

	client := w.getClient()
	files, err := client.ReadDir(prefix)
	if err != nil {
		return listing, err
	}
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			b := float64(file.Size())
			kb := b / 1024
			mb := kb / 1024
			current := filesystems.Listing{
				LastModified: file.ModTime(),
				Key:          file.Name(),
				Size:         mb,
				IsDir:        file.IsDir(),
			}
			listing = append(listing, current)
		}
	}

	return listing, nil
}

func (w *WebDAV) Delete(itemsToDelete []string) bool {
	client := w.getClient()
	for _, item := range itemsToDelete {
		err := client.Remove(item)
		if err != nil {
			return false
		}
	}
	return true
}

func (w *WebDAV) Get(destination string, items ...string) error {
	client := w.getClient()
	for _, item := range items {
		err := func() error {
			webdavFilePath := item
			localFilePath := fmt.Sprintf("%s/%s", destination, path.Base(item))

			reader, err := client.ReadStream(webdavFilePath)
			if err != nil {
				return err
			}

			file, err := os.Create(localFilePath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(file, reader)
			return err
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
