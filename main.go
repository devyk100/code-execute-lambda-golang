package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type RequestBody struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Inputs   string `json:"input"`
}

func pyExec(Code string, Inputs string) string {
	err := os.WriteFile("/tmp/main.py", []byte(Code), 0777)
	if err != nil {
		return err.Error()
	}
	cmd := exec.Command("python3", "/tmp/main.py")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err.Error()
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, Inputs)
	}()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output) + "\n" + err.Error()
	}
	fmt.Print(string(output))
	return string(output)
}

func nodeExec(Code string, Inputs string) string {
	err := os.WriteFile("/tmp/index.js", []byte(Code), 0777)
	if err != nil {
		return err.Error()
	}
	cmd := exec.Command("node", "/tmp/index.js")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err.Error()
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, Inputs)
	}()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output) + "\n" + err.Error()
	}
	return string(output)
}

func cppExec(Code string, Inputs string) string {
	err := os.WriteFile("/tmp/main.cpp", []byte(Code), 0777)
	if err != nil {
		return err.Error()
	}
	cmd := exec.Command("g++", "/tmp/main.cpp", "-o", "/tmp/a.out")
	cmd.Run()
	cmd2 := exec.Command("/tmp/a.out")
	stdin, err := cmd2.StdinPipe()
	if err != nil {
		return err.Error()
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, Inputs)
	}()
	output2, err2 := cmd2.CombinedOutput()
	if err2 != nil {
		return string(output2) + "\n" + err2.Error()
	}
	return string(output2)
}

func javaExec(Code string, Inputs string) string {
	err := os.WriteFile("/tmp/main.java", []byte(Code), 0755)
	if err != nil {
		return err.Error()
	}
	cmd := exec.Command("java", "/tmp/main.java")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err.Error()
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, Inputs)
	}()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output) + "\n" + err.Error()
	}
	return string(output)
}

func codeExecutor(Language string, Code string, output chan string, Inputs string) {
	if Language == "python" {
		output <- pyExec(Code, Inputs)
	} else if Language == "java" {
		output <- javaExec(Code, Inputs)
	} else if Language == "c++" {
		output <- cppExec(Code, Inputs)
	} else if Language == "node.js" {
		output <- nodeExec(Code, Inputs)
	} else {
		output <- "Invalid input"
	}
	// return "invalid"
}

func timeOut(sec float32, output chan string) {
	time.Sleep(time.Duration(sec) * time.Second)
	output <- "Operation timed out"
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var requestBody RequestBody

	// Parse the JSON body into the struct
	err := json.Unmarshal([]byte(event.Body), &requestBody)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "some error",
		}, err
	}

	outputChan := make(chan string)
	go codeExecutor(requestBody.Language, requestBody.Code, outputChan, requestBody.Inputs)
	go timeOut(12, outputChan)
	output := <-outputChan
	var trimmedOutput string = ""
	byteSlice := []byte(output)
	if len(byteSlice) > 5000000 {
		trimmedOutput += "The output was trimmed:\n"
		trimmedOutput += string(byteSlice[0:5000000])
	} else {
		trimmedOutput = output
	}
	response := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       trimmedOutput,
	}
	return (response), nil
}

func main() {
	lambda.Start(handler)
}
