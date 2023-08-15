package openai

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var oai *OpenAI

func TestMain(m *testing.M) {
	godotenv.Load("../.env")
	openai, err := NewOpenAI()
	if err != nil {
		panic(err)
	}

	oai = openai

	os.Exit(m.Run())
}

func Test_NewOpenAI(t *testing.T) {
	_, err := NewOpenAI()
	if err != nil {
		t.Error(err)
	}

	key := os.Getenv("OPENAI_API_KEY")
	os.Setenv("OPENAI_API_KEY", "")

	_, err = NewOpenAI()
	if err == nil {
		t.Error("expected error")
	}

	os.Setenv("OPENAI_API_KEY", key)
}

// func Test_GPT3Dot5Turbo_ChatCompletion(t *testing.T) {
// 	prompt := "Could you explain to me how cold fusion works?"
// 	response, err := oai.GPT3Dot5Turbo_ChatCompletion(prompt, nil, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if response.Content == "" {
// 		t.Error("expected response content")
// 	}

// 	fmt.Printf("Response: %s\n", response.Content)
// }

// func Test_GPT4_ChatCompletion(t *testing.T) {
// 	prompt := "Could you explain to me how cold fusion works?"
// 	response, err := oai.GPT4_ChatCompletion(prompt, nil, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if response.Content == "" {
// 		t.Error("expected response content")
// 	}

// 	fmt.Printf("Response: %s\n", response.Content)
// }

// func Test_GPT3Dot5Turbo_StreamChatCompletion(t *testing.T) {
// 	prompt := "Could you explain to me how cold fusion works?"
// 	stream, err := oai.GPT3Dot5Turbo_StreamChatCompletion(prompt, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	defer stream.Close()

// 	fmt.Printf("Stream Response: ")
// 	for {
// 		response, err := stream.Recv()
// 		if errors.Is(err, io.EOF) {
// 			fmt.Println("\nStream Finished!")
// 			return
// 		}

// 		if err != nil {
// 			fmt.Printf("\nStream Error: %v\n", err)
// 			return
// 		}

// 		fmt.Printf(response.Choices[0].Delta.Content)
// 	}
// }

// func Test_GenerateImageReturningLinks(t *testing.T) {
// 	prompt := "A painting of a forest"
// 	imagelinks, err := oai.GenerateImageReturningLinks(prompt, 256, 1)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if len(imagelinks) == 0 {
// 		t.Error("expected image links")
// 	}

// 	for _, link := range imagelinks {
// 		fmt.Println(link)
// 	}
// }

// func Test_GenerateImageToFile(t *testing.T) {
// 	prompt := "A painting of a forest"
// 	imagelinks, err := oai.GenerateImageToFile(prompt, "./test", "painting", 256, 1)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if len(imagelinks) == 0 {
// 		t.Error("expected image links")
// 	}

// 	for _, link := range imagelinks {
// 		fmt.Println(link)
// 	}
// }
