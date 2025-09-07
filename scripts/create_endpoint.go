package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"SensorContinuum/pkg/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"

	"path/filepath"
)

const localstackEndpoint = "http://localhost:4566"

var envFilePath1 = filepath.Join("internal", "client", "environment", ".env")
var envFilePath2 = filepath.Join("site", ".env.development")

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
		// region endpoints
		{"regionList", "region"},
		{"regionSearchName", "region"},
		{"regionDataAggregated", "region"},

		// macrozone endpoints
		{"macrozoneList", "macrozone"},
		{"macrozoneSearchName", "macrozone"},

		{"macrozoneDataAggregatedName", "macrozone"},
		{"macrozoneDataAggregatedLocation", "macrozone"},

		{"macrozoneDataTrend", "macrozone"},
		{"macrozoneDataVariation", "macrozone"},
		{"macrozoneDataVariationCorrelation", "macrozone"},

		// zone endpoints
		{"zoneList", "zone"},
		{"zoneSearchName", "zone"},
		{"sensorDataRaw", "zone"},
		{"zoneDataAggregated", "zone"},
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

		var segments []string
		var envKey string
		switch lambda.Name {
		case "regionList":
			segments = []string{"list"}
			envKey = "REGION_LIST_URL"
		case "regionSearchName":
			segments = []string{"search", "name", "{name}"}
			envKey = "REGION_SEARCH_NAME_URL"
		case "regionDataAggregated":
			segments = []string{"data", "aggregated", "{region}"}
			envKey = "REGION_DATA_AGGREGATED_URL"

		case "macrozoneList":
			segments = []string{"list", "{region}"}
			envKey = "MACROZONE_LIST_URL"
		case "macrozoneSearchName":
			segments = []string{"search", "name", "{region}", "{name}"}
			envKey = "MACROZONE_SEARCH_NAME_URL"

		case "macrozoneDataAggregatedName":
			segments = []string{"data", "aggregated", "name", "{region}", "{macrozone}"}
			envKey = "MACROZONE_DATA_AGGREGATED_NAME_URL"
		case "macrozoneDataAggregatedLocation":
			segments = []string{"data", "aggregated", "location"}
			envKey = "MACROZONE_DATA_AGGREGATED_LOCATION_URL"
		case "macrozoneDataVariation":
			segments = []string{"data", "variation", "{region}"}
			envKey = "MACROZONE_DATA_VARIATION_URL"
		case "macrozoneDataVariationCorrelation":
			segments = []string{"data", "variation", "correlation", "{region}"}
			envKey = "MACROZONE_DATA_VARIATION_CORRELATION_URL"
		case "macrozoneDataTrend":
			segments = []string{"data", "trend", "{region}"}
			envKey = "MACROZONE_DATA_TREND_URL"

		case "zoneList":
			segments = []string{"list", "{region}", "{macrozone}"}
			envKey = "ZONE_LIST_URL"
		case "zoneSearchName":
			segments = []string{"search", "name", "{region}", "{macrozone}", "{name}"}
			envKey = "ZONE_SEARCH_NAME_URL"
		case "sensorDataRaw":
			segments = []string{"sensor", "data", "raw", "{region}", "{macrozone}", "{zone}", "{sensor}"}
			envKey = "ZONE_SENSOR_DATA_RAW_URL"
		case "zoneDataAggregated":
			segments = []string{"data", "aggregated", "{region}", "{macrozone}", "{zone}"}
			envKey = "ZONE_DATA_AGGREGATED_URL"
		default:
			segments = []string{lambda.Name}
			envKey = strings.ToUpper(lambda.Name) + "_URL"
		}

		// Crea risorse per ogni segmento del path
		parent := subfolderID
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

		lambdaArn := fmt.Sprintf("arn:cloudformation:lambda:us-east-1:000000000000:function:%s", lambda.Name)
		logger.Log.Info("Creo integrazione Lambda per: ", lambda.Name)
		_, err = apiClient.PutIntegration(ctx, &apigateway.PutIntegrationInput{
			RestApiId:             aws.String(apiID),
			ResourceId:            aws.String(resourceID),
			HttpMethod:            aws.String("GET"),
			Type:                  "AWS_PROXY",
			IntegrationHttpMethod: aws.String("POST"),
			Uri:                   aws.String(fmt.Sprintf("arn:cloudformation:apigateway:us-east-1:lambda:path/2015-03-31/functions/%s/invocations", lambdaArn)),
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
		resourcePath := lambda.Subfolder + "/" + strings.Join(segments, "/")
		url := fmt.Sprintf("http://localhost:4566/restapis/%s/dev/_user_request_/%s", apiID, resourcePath)
		envVars[envKey] = url
		logger.Log.Info("Endpoint creato: ", lambda.Name, " ", url)
	}

	// Scrivi tutte le variabili nel file .env.development
	f, err := os.Create(envFilePath1)
	if err != nil {
		logger.Log.Error("Impossibile creare il file .env: ", err)
		os.Exit(1)
	}
	defer f.Close()
	for k, v := range envVars {
		fmt.Fprintf(f, "%s=%s\n", k, v)
	}

	// Scrivi tutte le variabili nel file .env.development
	f, err = os.Create(envFilePath2)
	if err != nil {
		logger.Log.Error("Impossibile creare il file .env.development: ", err)
		os.Exit(1)
	}
	defer f.Close()
	for k, v := range envVars {
		fmt.Fprintf(f, "REACT_APP_%s=%s\n", k, strings.Replace(v, "http://localhost:4566", "", 1))
	}
}
