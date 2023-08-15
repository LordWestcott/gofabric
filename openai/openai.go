package openai

// https://github.com/sashabaranov/go-openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image/png"
	"os"
	"strconv"

	"github.com/sashabaranov/go-openai"
	ai "github.com/sashabaranov/go-openai"
)

type OpenAI struct {
	client *ai.Client
}

func NewOpenAI() (*OpenAI, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")

	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY not found")
	}

	client := ai.NewClient(apiKey)

	return &OpenAI{
		client: client,
	}, nil
}

func (o *OpenAI) GPT3Dot5Turbo_ChatCompletion(prompt string, previousMessages *[]ai.ChatCompletionMessage, functions *[]ai.FunctionDefinition) (*ai.ChatCompletionMessage, error) {

	if previousMessages == nil {
		previousMessages = &[]ai.ChatCompletionMessage{}
	}

	messages := append(*previousMessages, ai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	if functions == nil {
		functions = &[]ai.FunctionDefinition{}
	}
	f := *functions

	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		ai.ChatCompletionRequest{
			Model:     ai.GPT3Dot5Turbo,
			Messages:  messages,
			Functions: f,
		},
	)

	if err != nil {
		return nil, err
	}

	return &resp.Choices[0].Message, nil
}

// // This is not working yet
// func (o *OpenAI) GPT4_ChatCompletion(prompt string, previousMessages *[]ai.ChatCompletionMessage, functions *[]ai.FunctionDefinition) (*ai.ChatCompletionMessage, error) {

// 	if previousMessages == nil {
// 		previousMessages = &[]ai.ChatCompletionMessage{}
// 	}

// 	messages := append(*previousMessages, ai.ChatCompletionMessage{
// 		Role:    openai.ChatMessageRoleUser,
// 		Content: prompt,
// 	})

// 	if functions == nil {
// 		functions = &[]ai.FunctionDefinition{}
// 	}
// 	f := *functions

// 	resp, err := o.client.CreateChatCompletion(
// 		context.Background(),
// 		ai.ChatCompletionRequest{
// 			Model:     ai.GPT4,
// 			Messages:  messages,
// 			Functions: f,
// 		},
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return &resp.Choices[0].Message, nil
// }

//To use the stream:
/*
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}
*/
func (o *OpenAI) GPT3Dot5Turbo_StreamChatCompletion(prompt string, previousMessages *[]ai.ChatCompletionMessage) (*ai.ChatCompletionStream, error) {

	if previousMessages == nil {
		previousMessages = &[]ai.ChatCompletionMessage{}
	}

	messages := append(*previousMessages, ai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	req := ai.ChatCompletionRequest{
		Model:    ai.GPT3Dot5Turbo,
		Messages: messages,
		Stream:   true,
	}
	ctx := context.Background()

	stream, err := o.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

// // This is not working yet
//To use the stream:
/*
	defer stream.Close()

	fmt.Printf("Stream response: ")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\nStream finished")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}
*/
// func (o *OpenAI) GPT4_StreamChatCompletion(prompt string, previousMessages *[]ai.ChatCompletionMessage) (*ai.ChatCompletionStream, error) {

// 	if previousMessages == nil {
// 		previousMessages = &[]ai.ChatCompletionMessage{}
// 	}

// 	messages := append(*previousMessages, ai.ChatCompletionMessage{
// 		Role:    openai.ChatMessageRoleUser,
// 		Content: prompt,
// 	})

// 	req := ai.ChatCompletionRequest{
// 		Model:    ai.GPT4,
// 		Messages: messages,
// 		Stream:   true,
// 	}
// 	ctx := context.Background()

// 	stream, err := o.client.CreateChatCompletionStream(ctx, req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return stream, nil
// }

func (o *OpenAI) AudioFileToText(audioFilePath string) (string, error) {
	ctx := context.Background()

	req := openai.AudioRequest{
		Model:    ai.Whisper1,
		FilePath: audioFilePath,
	}

	resp, err := o.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}

// size: 256, 512, 1024
func (o *OpenAI) GenerateImageReturningLinks(prompt string, size, amount int) ([]string, error) {
	ctx := context.Background()
	images := []string{}

	s := ""
	switch size {
	case 256:
		s = ai.CreateImageSize256x256
	case 512:
		s = ai.CreateImageSize512x512
	case 1024:
		s = ai.CreateImageSize1024x1024
	default:
		return images, errors.New("size not supported, please use 256, 512 or 1024")
	}

	req := openai.ImageRequest{
		Prompt:         prompt,
		Size:           s,
		ResponseFormat: ai.CreateImageResponseFormatURL,
		N:              amount,
	}

	resp, err := o.client.CreateImage(ctx, req)
	if err != nil {
		return images, err
	}

	for _, image := range resp.Data {
		images = append(images, image.URL)
	}

	return images, nil
}

func (o *OpenAI) GenerateImageToFile(prompt, filePath, fileName string, size, amount int) ([]string, error) {
	ctx := context.Background()
	images := []string{}

	s := ""
	switch size {
	case 256:
		s = ai.CreateImageSize256x256
	case 512:
		s = ai.CreateImageSize512x512
	case 1024:
		s = ai.CreateImageSize1024x1024
	default:
		return images, errors.New("size not supported, please use 256, 512 or 1024")
	}

	req := openai.ImageRequest{
		Prompt:         prompt,
		Size:           s,
		ResponseFormat: ai.CreateImageResponseFormatB64JSON,
		N:              amount,
	}

	resp, err := o.client.CreateImage(ctx, req)
	if err != nil {
		return images, err
	}

	for i, image := range resp.Data {
		imgBytes, err := base64.StdEncoding.DecodeString(image.B64JSON)
		if err != nil {
			return images, err
		}

		r := bytes.NewReader(imgBytes)
		imgData, err := png.Decode(r)
		if err != nil {
			return images, err
		}

		resolvedFilePath := filePath + "/" + fileName + "-" + strconv.Itoa(i) + ".png"
		file, err := os.Create(resolvedFilePath)
		if err != nil {
			return images, err
		}
		defer file.Close()

		if err := png.Encode(file, imgData); err != nil {
			return images, err
		}

		images = append(images, resolvedFilePath)
	}

	return images, nil
}
