package miniofilesystem

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/Env-Co-Ltd/framinGo/filesystems"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Minio is the overall type for the minio filesystem, and contains
// the connection credentials, endpoint, and the bucket to use
type Minio struct {
	Endpoint string
	Key      string
	Secret   string
	UseSSL   bool
	Region   string
	Bucket   string
}

// getCredentials generates a minio client using the credentials stored in
// the Minio type
func (m *Minio) getCredentials() *minio.Client {
	log.Printf("Connecting to MinIO at %s with bucket %s", m.Endpoint, m.Bucket)
	client, err := minio.New(m.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.Key, m.Secret, ""),
		Secure: m.UseSSL,
	})
	if err != nil {
		log.Printf("Error creating MinIO client: %v", err)
		return nil
	}
	return client
}

// Put transfers a file to the remote file system
func (m *Minio) Put(fileName, folder string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	objectName := path.Base(fileName)
	client := m.getCredentials()
	uploadInfo, err := client.FPutObject(ctx, m.Bucket, fmt.Sprintf("%s/%s", folder, objectName), fileName, minio.PutObjectOptions{})
	if err != nil {
		log.Println("Failed with FPutObject")
		log.Println(err)
		log.Println("UploadInfo:", uploadInfo)
		return err
	}

	return nil
}

// List returns a listing of all files in the remote bucket with the
// given prefix, except for files with a leading . in the name
func (m *Minio) List(prefix string) ([]filesystems.Listing, error) {
	var listing []filesystems.Listing

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := m.getCredentials()

	// 先頭のスラッシュを削除
	prefix = strings.TrimPrefix(prefix, "/")
	log.Printf("Listing objects with prefix: '%s'", prefix)

	// バケットが存在するか確認
	exists, err := client.BucketExists(ctx, m.Bucket)
	if err != nil {
		log.Printf("Error checking bucket: %v", err)
		return nil, err
	}
	if !exists {
		log.Printf("Bucket %s does not exist", m.Bucket)
		return nil, fmt.Errorf("bucket %s does not exist", m.Bucket)
	}

	objectCh := client.ListObjects(ctx, m.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Error listing objects: %v", object.Err)
			return listing, object.Err
		}

		log.Printf("Found object: Key='%s', Size=%d bytes", object.Key, object.Size)
		if !strings.HasPrefix(object.Key, ".") {
			b := float64(object.Size)
			kb := b / 1024
			mb := kb / 1024
			item := filesystems.Listing{
				Etag:         object.ETag,
				LastModified: object.LastModified,
				Key:          object.Key,
				Size:         mb,
			}
			listing = append(listing, item)
			log.Printf("Added to listing: %s", object.Key)
		}
	}

	log.Printf("Total objects found: %d", len(listing))
	return listing, nil
}

// Delete removes one or more files from the remote filesystem
func (m *Minio) Delete(itemsToDelete []string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := m.getCredentials()

	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
	}

	for _, item := range itemsToDelete {
		err := client.RemoveObject(ctx, m.Bucket, item, opts)
		if err != nil {
			fmt.Println(err)
			return false
		}
	}
	return true
}

// Get pulls a file from the remote file system and saves it somewhere on our server
func (m *Minio) Get(destination string, items ...string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := m.getCredentials()

	for _, item := range items {
		err := client.FGetObject(ctx, m.Bucket, item, fmt.Sprintf("%s/%s", destination, path.Base(item)), minio.GetObjectOptions{})
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}
