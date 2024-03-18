package clients

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
)

func NewDynamoDBClient() *dynamodb.DynamoDB {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String("http://localhost:4566"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
	})
	if err != nil {
		log.Fatalf("Failed to create session for DynamoDB: %v", err)
	}

	return dynamodb.New(sess)
}

func NewElasticsearchClient(domainName string) *elasticsearch.Client {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String("http://localhost:4566"),
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
	})
	if err != nil {
		log.Fatalf("Failed to create session for Elasticsearch: %v", err)
	}

	esSvc := elasticsearchservice.New(sess)
	describeParams := &elasticsearchservice.DescribeElasticsearchDomainInput{
		DomainName: aws.String(domainName),
	}
	describeResp, err := esSvc.DescribeElasticsearchDomain(describeParams)
	if err != nil {
		log.Fatalf("Failed to describe Elasticsearch domain: %v", err)
	}

	esEndpoint := describeResp.DomainStatus.Endpoint
	if esEndpoint == nil {
		log.Fatalf("Elasticsearch domain endpoint is nil")
	}

	esConfig := elasticsearch.Config{
		Addresses: []string{fmt.Sprintf("http://%s", *esEndpoint)},
	}
	esClient, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
	}

	return esClient
}
