// Description: This file contains the functions to interact with Google Cloud Storage.
package cmd

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"io"
	"os"
)

// deleteFile deletes a file from Google Cloud Storage
func deleteFile(client *storage.Client, bucketName, objectName string) error {
	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)

	return object.Delete(context.Background())
}

// uploadFile uploads a file to Google Cloud Storage
// document this function with comments
func uploadFile(client *storage.Client, bucketName, objectName, filePath string) (string, error) {
	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)
	newCtx := context.Background()
	wc := object.NewWriter(newCtx)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	if _, err := io.Copy(wc, file); err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Return the public URL of the uploaded object.
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)

	fmt.Printf("File %s uploaded to bucket %s as %s and is available at %s.\n", filePath, bucketName, objectName, publicURL)
	return publicURL, nil
}
