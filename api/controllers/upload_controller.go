package controllers

import (
	"bytes"
	"fmt"
	"handCup-project-backend/api/responses"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/globalsign/mgo/bson"
	"github.com/joho/godotenv"
)

func (server *Server) UploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `file`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("Error Retrieving the File")
		log.Println(err)
		return
	}
	defer file.Close()
	log.Println("Uploaded File: ", handler.Filename)
	log.Println("File Size: ", handler.Size)
	log.Println("MIME Header: ", handler.Header)

	// // Create a temporary file within our temp-images directory that follows
	// // a particular naming pattern
	// tempFile, err := ioutil.TempFile("temp-images", "upload-*.png")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// defer tempFile.Close()

	// // read all of the contents of our uploaded file into a
	// // byte array
	// fileBytes, err := ioutil.ReadAll(file)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// // write this byte array to our temporary file
	// tempFile.Write(fileBytes)
	// // return that we have successfully uploaded our file!
	// fmt.Println(w, "Successfully Uploaded File\n")

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not comming through %v", err)
	} else {
		log.Println("We are getting the env values")
	}
	// create an AWS session which can be
	// reused if we're uploading many files
	s, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_S3_KEY_ID"),     // id
			os.Getenv("AWS_S3_SECRET_KEY"), // secret
			""),                            // token can be left blank for now
	})
	if err != nil {
		fmt.Fprintf(w, "Could not create AWS session")
	}
	fileName, err := UploadFileToS3(s, file, handler)
	if err != nil {
		fmt.Fprintf(w, "Could not upload file to S3")
	}

	log.Println(w, "Image uploaded successfully:", fileName)
	url := os.Getenv("AWS_S3_URL") + fileName
	responses.JSON(w, http.StatusOK, url)
}

// UploadFileToS3 saves a file to aws bucket and returns the url to // the file and an error if there's any
func UploadFileToS3(s *session.Session, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// get the file size and read
	// the file content into a buffer
	size := fileHeader.Size
	buffer := make([]byte, size)
	file.Read(buffer)

	// create a unique file name for the file
	tempFileName := "media/" + bson.NewObjectId().Hex() + filepath.Ext(fileHeader.Filename)

	// config settings: this is where you choose the bucket,
	// filename, content-type and storage class of the file
	// you're uploading
	_, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String("boba-maps"),
		Key:                  aws.String(tempFileName),
		ACL:                  aws.String("public-read"), // could be private if you want it to be access by only authorized users
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(int64(size)),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		// StorageClass: aws.String("INTELLIGENT_TIERING"),
	})
	if err != nil {
		return "", err
	}

	return tempFileName, err
}
