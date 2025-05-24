package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

// TODO: move to env
const BUCKET_NAME = "files-test-bucket"
const DEST_DIR = "/Users/vbortniak/Projects/file-syncer/test-folder"

func DownloadFile(objectName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Create the full destination path
	destPath := filepath.Join(DEST_DIR, filepath.Base(objectName))

	// Create the destination file
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("os.Create: %v", err)
	}
	defer f.Close()

	// Get object from bucket
	rc, err := client.Bucket(BUCKET_NAME).Object(objectName).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("Object(%q).NewReader: %v", objectName, err)
	}
	defer rc.Close()

	// Copy data to local file
	if _, err := io.Copy(f, rc); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	fmt.Printf("File %v downloaded to %v\n", objectName, destPath)
	return nil
}

func UploadFile(path string) {
	uploadFileToStorage(BUCKET_NAME, path)
}
func DeleteFile(path string) {
	safeDelete(BUCKET_NAME, path)
}

// safeDelete checks if file exists before deleting
func safeDelete(bucketName, path string) error {
	objectName := getFileNameFromPath(path)

	exists, err := checkFileExists(bucketName, objectName)
	if err != nil {
		return fmt.Errorf("error checking file existence: %v", err)
	}

	if !exists {
		log.Printf("File %s does not exist in bucket %s\n", objectName, bucketName)
		return nil
	}

	return deleteFileFromStorage(bucketName, objectName)
}

// checkFileExists checks if a file exists before attempting deletion
func checkFileExists(bucketName, objectName string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return false, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	obj := client.Bucket(bucketName).Object(objectName)
	_, err = obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("Object(%q).Attrs: %v", objectName, err)
	}
	return true, nil
}

// uploadFile uploads an object.
func uploadFileToStorage(bucketName, filePath string) error {
	log.Println("inside Upload file to bucket in this block")
	objectName := getFileNameFromPath(filePath)
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()

	log.Println("objectName = ", objectName)

	// Create a client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Open the local file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("os.Open: %v", err)
	}
	defer file.Close()

	// Get bucket handle
	bucket := client.Bucket(bucketName)

	// Create object handle and writer
	obj := bucket.Object(objectName)
	writer := obj.NewWriter(ctx)

	// Set optional attributes
	writer.ContentType = "application/octet-stream" // Set appropriate content type
	writer.Metadata = map[string]string{
		"uploaded-by": "go-example",
		"upload-time": time.Now().Format(time.RFC3339),
	}

	// Copy file content to GCS
	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}

	// Close the writer to finalize the upload
	if err := writer.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	log.Printf("File %s uploaded to bucket %s as %s\n", filePath, bucketName, objectName)
	return nil
}

// deleteFile deletes a single file from GCS
func deleteFileFromStorage(bucketName, objectName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Get object handle and delete
	obj := client.Bucket(bucketName).Object(objectName)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %v", objectName, err)
	}

	log.Printf("File %s deleted from bucket %s\n", objectName, bucketName)
	return nil

}

func getFileNameFromPath(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}
