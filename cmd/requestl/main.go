package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/etowett/requestl/build"
)

var (
	AwsSession *session.Session
)

type QueueRequest struct {
	Count int `json:"count"`
}

func main() {
	logFile := os.Getenv("LOG_FILE")
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		wrt := io.MultiWriter(os.Stdout, f)
		log.SetOutput(wrt)
	}

	AwsSession = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	go processQueueMessages()

	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/queue", queueStuff)
	http.HandleFunc("/", handleRequest)

	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "7000"
	}

	log.Printf("Server starting, listening on :%v", serverPort)
	http.ListenAndServe(fmt.Sprintf(":%v", serverPort), nil)
}

func printRequest(r *http.Request) error {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err
	}

	log.Printf("\nBody:\n %+v \n", string(dump))
	log.Printf("\n===================================================\n")
	return nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	err := printRequest(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error dumping request: %v", err), http.StatusInternalServerError)
		return
	}

	theResponse := map[string]interface{}{
		"success": true,
		"status":  "Ok",
	}

	jsResp, err := json.Marshal(theResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsResp)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	err := printRequest(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error dumping request: %v", err), http.StatusInternalServerError)
		return
	}

	theResponse := map[string]interface{}{
		"success":    true,
		"status":     "Ok",
		"sha1ver":    build.Sha1Ver,
		"build_time": build.Time,
		"git_commit": build.GitCommit,
		"git_branch": build.GitBranch,
		"version":    build.Version,
	}

	jsResp, err := json.Marshal(theResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsResp)
}

func getQueueURL(
	sess *session.Session,
	queue string,
) (*sqs.GetQueueUrlOutput, error) {
	// Create an SQS service client
	svc := sqs.New(sess)

	result, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queue,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func pushMessageToQueue(
	sess *session.Session,
	queueURL *string,
	data map[string]interface{},
) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// Create an SQS service client
	// snippet-start:[sqs.go.send_message.call]
	svc := sqs.New(sess)

	_, err = svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"BinaryValue": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String("The Whistler"),
			},
			"Author": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String("John Grisham"),
			},
			"WeeksOn": &sqs.MessageAttributeValue{
				DataType:    aws.String("Number"),
				StringValue: aws.String("6"),
			},
			"User": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String("Eutychus Towett"),
			},
		},
		MessageBody: aws.String(string(jsonData)),
		QueueUrl:    queueURL,
	})
	// snippet-end:[sqs.go.send_message.call]
	if err != nil {
		return err
	}

	return nil
}

func queueStuff(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var data QueueRequest
	err := decoder.Decode(&data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not decode request: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("data: =[%+v]", data)

	result, err := getQueueURL(AwsSession, os.Getenv("DATA_QUEUE"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not get queue url: %v", err), http.StatusInternalServerError)
		return
	}

	for i := 0; i < data.Count; i++ {
		queueReq := map[string]interface{}{
			"index":   i,
			"message": fmt.Sprintf("message at %v", time.Now().String()),
		}
		log.Printf("queuing: =[%+v]", queueReq)

		err = pushMessageToQueue(AwsSession, result.QueueUrl, queueReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could queue data: %v", err), http.StatusInternalServerError)
			return
		}
	}

	theResponse := map[string]interface{}{
		"success": true,
		"status":  "Ok",
		"time":    time.Now().String(),
	}

	jsResp, err := json.Marshal(theResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsResp)
}

func getMessages(
	sess *session.Session,
	queueURL *string,
	timeout *int64,
) (*sqs.ReceiveMessageOutput, error) {
	// Create an SQS service client
	svc := sqs.New(sess)

	// snippet-start:[sqs.go.receive_messages.call]
	msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            queueURL,
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   timeout,
	})
	// snippet-end:[sqs.go.receive_messages.call]
	if err != nil {
		return nil, err
	}

	return msgResult, nil
}

func processQueueMessages() {
	result, err := getQueueURL(AwsSession, os.Getenv("DATA_QUEUE"))
	if err != nil {
		log.Printf("error getting queue url: %+v", err)
		return
	}

	// timeout := int64(5)
	// msgResult, err := getMessages(AwsSession, result.QueueUrl, &timeout)
	// if err != nil {
	// 	log.Printf("error getting message: %+v", err)
	// 	return
	// }
	// log.Printf("msgResult: =[%+v]", msgResult)

	checkMessages(AwsSession, result.QueueUrl)
}

func checkMessages(sess *session.Session, queueURL *string) {
	sqsSvc := sqs.New(sess)

	for {
		retrieveMessageRequest := sqs.ReceiveMessageInput{
			QueueUrl: queueURL,
		}

		retrieveMessageResponse, _ := sqsSvc.ReceiveMessage(&retrieveMessageRequest)

		if len(retrieveMessageResponse.Messages) > 0 {

			processedReceiptHandles := make([]*sqs.DeleteMessageBatchRequestEntry, len(retrieveMessageResponse.Messages))

			for i, mess := range retrieveMessageResponse.Messages {
				log.Printf("message =[%v] == %+v", i, mess.String())

				processedReceiptHandles[i] = &sqs.DeleteMessageBatchRequestEntry{
					Id:            mess.MessageId,
					ReceiptHandle: mess.ReceiptHandle,
				}
			}

			deleteMessageRequest := sqs.DeleteMessageBatchInput{
				QueueUrl: queueURL,
				Entries:  processedReceiptHandles,
			}

			_, err := sqsSvc.DeleteMessageBatch(&deleteMessageRequest)

			if err != nil {
				log.Print("Failed to delete batch: %v", err)
			}
		}

		if len(retrieveMessageResponse.Messages) == 0 {
			log.Printf(":(  I have no messages")
			log.Printf("%v", time.Now())
			time.Sleep(time.Second * 15)
		}
	}
}
