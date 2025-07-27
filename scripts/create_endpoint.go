package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"SensorContinuum/pkg/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	apigateway "github.com/aws/aws-sdk-go-v2/service/apigateway"
)

const localstackEndpoint = "http://localhost:4566"
const envFilePath = "../internal/client/environment/.env"

type LambdaFunc struct {
	Name      string
	Subfolder string
}

func main() {
	logger.CreateLogger(logger.Context{
		"service": "api-backend",
		"module":  "create_endpoint",
	})

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           localstackEndpoint,
					SigningRegion: "us-east-1",
				}, nil
			}),
		),
	)
	if err != nil {
		logger.Log.Error("Impossibile caricare la configurazione AWS SDK: ", err)
		os.Exit(1)
	}
	apiClient := apigateway.NewFromConfig(cfg)

	lambdas := []LambdaFunc{
		{"regionList", "region"},
		{"regionSearchId", "region"},
		{"regionSearchName", "region"},
		{"buildingList", "building"},
		{"buildingSearchId", "building"},
		{"buildingSearchName", "building"},
		{"buildingSearchRegion", "building"},
	}

	apiIDs := make(map[string]string)
	subfolderIDs := make(map[string]string)
	envVars := make(map[string]string)

	for _, lambda := range lambdas {
		apiID := apiIDs[lambda.Subfolder]
		var subfolderID string

		if apiID == "" {
			logger.Log.Info("Creo REST API per subfolder: ", lambda.Subfolder)
			apiOut, err := apiClient.CreateRestApi(ctx, &apigateway.CreateRestApiInput{
				Name: aws.String(lambda.Subfolder + "-api"),
			})
			if err != nil {
				logger.Log.Error("Errore CreateRestApi: ", err)
				os.Exit(1)
			}
			apiID = *apiOut.Id
			apiIDs[lambda.Subfolder] = apiID

			resList, err := apiClient.GetResources(ctx, &apigateway.GetResourcesInput{
				RestApiId: aws.String(apiID),
			})
			if err != nil {
				logger.Log.Error("Errore GetResources: ", err)
				os.Exit(1)
			}
			var parentID string
			for _, r := range resList.Items {
				if *r.Path == "/" {
					parentID = *r.Id
					break
				}
			}
			logger.Log.Info("Creo risorsa subfolder: ", lambda.Subfolder)
			subRes, err := apiClient.CreateResource(ctx, &apigateway.CreateResourceInput{
				RestApiId: aws.String(apiID),
				ParentId:  aws.String(parentID),
				PathPart:  aws.String(lambda.Subfolder),
			})
			if err != nil {
				logger.Log.Error("Errore CreateResource subfolder: ", err)
				os.Exit(1)
			}
			subfolderID = *subRes.Id
			subfolderIDs[lambda.Subfolder] = subfolderID
		} else {
			subfolderID = subfolderIDs[lambda.Subfolder]
		}

		var pathPart string
		var param string
		var envKey string
		switch lambda.Name {
		case "regionList":
			pathPart = "list"
			envKey = "REGION_LIST_URL"
		case "buildingList":
			pathPart = "list"
			envKey = "BUILDING_LIST_URL"
		case "regionSearchId":
			pathPart = "search/id"
			param = "{id}"
			envKey = "REGION_SEARCH_ID_URL"
		case "buildingSearchId":
			pathPart = "search/id"
			param = "{id}"
			envKey = "BUILDING_SEARCH_ID_URL"
		case "regionSearchName":
			pathPart = "search/name"
			param = "{name}"
			envKey = "REGION_SEARCH_NAME_URL"
		case "buildingSearchName":
			pathPart = "search/name"
			param = "{name}"
			envKey = "BUILDING_SEARCH_NAME_URL"
		case "buildingSearchRegion":
			pathPart = "search/region"
			param = "{region_id}"
			envKey = "BUILDING_SEARCH_REGION_URL"
		default:
			pathPart = lambda.Name
			envKey = strings.ToUpper(lambda.Name) + "_URL"
		}

		// Crea risorse per ogni segmento del path
		parent := subfolderID
		segments := strings.Split(pathPart, "/")
		for _, seg := range segments {
			logger.Log.Info("Creo risorsa path: ", seg)
			res, err := apiClient.CreateResource(ctx, &apigateway.CreateResourceInput{
				RestApiId: aws.String(apiID),
				ParentId:  aws.String(parent),
				PathPart:  aws.String(seg),
			})
			if err != nil {
				logger.Log.Error("Errore CreateResource path: ", err)
				os.Exit(1)
			}
			parent = *res.Id
		}
		resourceID := parent

		// Se c'è un parametro, crea la risorsa parametro
		if param != "" {
			logger.Log.Info("Creo risorsa parametro: ", param)
			res2, err := apiClient.CreateResource(ctx, &apigateway.CreateResourceInput{
				RestApiId: aws.String(apiID),
				ParentId:  aws.String(resourceID),
				PathPart:  aws.String(param),
			})
			if err != nil {
				logger.Log.Error("Errore CreateResource param: ", err)
				os.Exit(1)
			}
			resourceID = *res2.Id
		}

		logger.Log.Info("Creo metodo GET su risorsa: ", lambda.Name)
		_, err = apiClient.PutMethod(ctx, &apigateway.PutMethodInput{
			RestApiId:         aws.String(apiID),
			ResourceId:        aws.String(resourceID),
			HttpMethod:        aws.String("GET"),
			AuthorizationType: aws.String("NONE"),
		})
		if err != nil {
			logger.Log.Error("Errore PutMethod: ", err)
			os.Exit(1)
		}

		lambdaArn := fmt.Sprintf("arn:aws:lambda:us-east-1:000000000000:function:%s", lambda.Name)
		logger.Log.Info("Creo integrazione Lambda per: ", lambda.Name)
		_, err = apiClient.PutIntegration(ctx, &apigateway.PutIntegrationInput{
			RestApiId:             aws.String(apiID),
			ResourceId:            aws.String(resourceID),
			HttpMethod:            aws.String("GET"),
			Type:                  "AWS_PROXY",
			IntegrationHttpMethod: aws.String("POST"),
			Uri:                   aws.String(fmt.Sprintf("arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/%s/invocations", lambdaArn)),
		})
		if err != nil {
			logger.Log.Error("Errore PutIntegration: ", err)
			os.Exit(1)
		}

		logger.Log.Info("Deploy API: ", apiID)
		_, err = apiClient.CreateDeployment(ctx, &apigateway.CreateDeploymentInput{
			RestApiId: aws.String(apiID),
			StageName: aws.String("dev"),
		})
		if err != nil {
			logger.Log.Error("Errore CreateDeployment: ", err)
			os.Exit(1)
		}

		// Costruisci il path finale per la stampa
		resourcePath := lambda.Subfolder + "/" + strings.ReplaceAll(pathPart, "/", "/")
		if param != "" {
			resourcePath += "/" + param
		}
		url := fmt.Sprintf("http://localhost:4566/restapis/%s/dev/_user_request_/%s", apiID, resourcePath)
		envVars[envKey] = url
		logger.Log.Info("Endpoint creato: ", lambda.Name, " ", url)
	}

	// Scrivi tutte le variabili nel file .env
	f, err := os.Create(envFilePath)
	if err != nil {
		logger.Log.Error("Impossibile creare il file .env: ", err)
		os.Exit(1)
	}
	defer f.Close()
	for k, v := range envVars {
		fmt.Fprintf(f, "%s=%s\n", k, v)
	}
}
