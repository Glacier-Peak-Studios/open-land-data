package main

import (
	"flag"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/schollz/progressbar/v3"
	utils "glacierpeak.app/openland/pkg/utils"
)

func main() {
	workersOpt := flag.Int("t", 1, "The number of concurrent jobs being processed")
	uploadDir := flag.String("i", "", "Directory to upload files from")
	region := flag.String("r", "us-west-2", "Region of the s3 bucket")
	bucket := flag.String("b", "gp.us.general", "Name of the bucket to upload tiles to")
	// zRange := flag.String("z", "18", "Zoom levels to generate. (Ex. \"2-16\") Must start with current zoom level")
	verboseOpt := flag.Int("v", 1, "Set the verbosity level:\n"+
		" 0 - Only prints error messages\n"+
		" 1 - Adds run specs and error details\n"+
		" 2 - Adds general progress info\n"+
		" 3 - Adds debug info and details more detail\n")
	flag.Parse()

	switch *verboseOpt {
	case 0:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 3:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	default:
		break
	}

	// sess, err := session.NewSession(&aws.Config{
	// 	Region: aws.String(*region)},
	// )

	// if err != nil {
	// 	exitErrorf("Couldn't create session", err)
	// }

	// uploader := s3manager.NewUploader(sess)

	// toUploadPipe := make(chan string, 500)
	// results := make(chan string, 1000)

	// go utils.GetAllTilesStreamed(*uploadDir, *workersOpt, toUploadPipe)

	// var workersDone uint64
	// workersTotal := uint64(*workersOpt)

	// for i := 0; i < *workersOpt; i++ {
	// 	go uploadTileBucketWorker(toUploadPipe, results, uploader, *bucket, &workersDone, workersTotal)
	// }

	// progBar := progressbar.NewOptions(-1,
	// 	progressbar.OptionSetDescription("Uploading files"),
	// 	progressbar.OptionSetItsString("files"),
	// 	progressbar.OptionShowIts(),
	// 	progressbar.OptionSpinnerType(14),
	// )

	// for result := range results {
	// 	log.Debug().Msg(result)
	// 	progBar.Add(1)
	// 	// println(result)
	// 	// println("rslt len:", len(results))

	// }
	// progBar.Finish()
	S3Upload(*region, *bucket, *uploadDir, *workersOpt)

}

func S3Upload(region string, bucket string, uploadDir string, workers int) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	if err != nil {
		exitErrorf("Couldn't create session", err)
	}

	uploader := s3manager.NewUploader(sess)

	toUploadPipe := make(chan string, 500)
	results := make(chan string, 1000)

	go utils.GetAllTilesStreamed(uploadDir, workers, toUploadPipe)

	var workersDone uint64
	workersTotal := uint64(workers)

	for i := 0; i < workers; i++ {
		go uploadTileBucketWorker(toUploadPipe, results, uploader, bucket, &workersDone, workersTotal)
	}

	progBar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("Uploading files"),
		progressbar.OptionSetItsString("files"),
		progressbar.OptionShowIts(),
		progressbar.OptionSpinnerType(14),
	)

	for result := range results {
		log.Debug().Msg(result)
		progBar.Add(1)
		// println(result)
		// println("rslt len:", len(results))

	}
	progBar.Finish()
}

func uploadTileBucketWorker(jobs <-chan string, results chan<- string, uploader *s3manager.Uploader, bucketName string, workersDone *uint64, workerCount uint64) {
	// ignoreZL := make(map[int]bool)
	// for i := 10; i <= 16; i++ {
	// 	ignoreZL[i] = true
	// }

	for job := range jobs {
		tile, _ := utils.PathToTile(job)

		// shouldUpload := !ignoreZL[tile.Z]
		shouldUpload := true

		// if tile.Z == 17 {
		// 	if tile.X <= 26300 {
		// 		shouldUpload = false
		// 	}
		// }

		if shouldUpload {
			err := uploadFileToBucket(uploader, bucketName, tile.GetPath(), job)
			if err != nil {
				results <- "Err: " + err.Error()
			} else {
				results <- "File done: " + job
			}
		} else {
			results <- "Skipping, excluded zoom: " + job
		}
	}
	atomic.AddUint64(workersDone, 1)
	if atomic.LoadUint64(workersDone) == workerCount {
		close(results)
	}
}

// func BucketFileUploadWorker(jobs)

func uploadFileToBucket(uploader *s3manager.Uploader, bucketName string, objectKey string, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		exitErrorf("Unable to open file %q, %v", err)
	}
	defer file.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		// Print the error and exit.
		// fmt.Printf("Unable to upload %q to %q, %v", filepath, bucketName, err)
		return errors.Errorf("Unable to upload %q to %q, %v", filepath, bucketName, err)
	}

	// fmt.Printf("Successfully uploaded %q to %q\n", filepath, bucketName)
	return nil
}

func printBucketItems(svc *s3.S3, bucketName string) {
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucketName)})
	if err != nil {
		exitErrorf("Unable to list items in bucket %q, %v", bucketName, err)
	}

	for _, item := range resp.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}
}

func printBuckets(svc *s3.S3) {
	result, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
