package yt

// import (
// 	"context"
// 	"fmt"
// 	"os"

// 	"google.golang.org/api/option"
// 	"google.golang.org/api/youtube/v3"
// )

// func (y *Youtube) Upload(v *Video) (string, error) {
// 	if v.filename == "" {
// 		return "", fmt.Errorf("filename is empty")
// 	}

// 	client := getClient(youtube.YoutubeUploadScope)

// 	ctx := context.Background()
// 	service, err := youtube.NewService(ctx, option.)
// 	if err != nil {
// 		return "", fmt.Errorf("Error creating YouTube client: %v", err)
// 	}

// 	upload := &youtube.Video{
// 		Snippet: &youtube.VideoSnippet{
// 			Title:       v.title,
// 			Description: v.description,
// 			CategoryId:  v.category,
// 			Tags:        v.keywords,
// 		},
// 		Status: &youtube.VideoStatus{PrivacyStatus: v.privacy},
// 	}

// 	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
// 	file, err := os.Open(v.filename)
// 	if err != nil {
// 		return "", fmt.Errorf("Error opening %v: %v", v.filename, err)
// 	}
// 	defer file.Close()

// 	response, err := call.Media(file).Do()
// 	if err != nil {
// 		return "", fmt.Errorf("Error making YouTube API call: %v", err)
// 	}
// 	fmt.Printf("Upload successful! Video ID: %v\n", response.Id)
// 	return response.Id, nil

// 	//Add Thumbnails.

// }
