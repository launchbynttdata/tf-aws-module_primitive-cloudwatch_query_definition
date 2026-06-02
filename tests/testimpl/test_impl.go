package testimpl

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/launchbynttdata/lcaf-component-terratest/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getCloudWatchLogsClient(t *testing.T, region string) *cloudwatchlogs.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	require.NoError(t, err, "unable to load AWS config")
	return cloudwatchlogs.NewFromConfig(cfg)
}

func findQueryDefinitionByName(t *testing.T, client *cloudwatchlogs.Client, name string) *cwltypes.QueryDefinition {
	t.Helper()
	var nextToken *string
	for {
		output, err := client.DescribeQueryDefinitions(context.TODO(), &cloudwatchlogs.DescribeQueryDefinitionsInput{
			QueryDefinitionNamePrefix: aws.String(name),
			NextToken:                 nextToken,
		})
		require.NoError(t, err, "DescribeQueryDefinitions should succeed")

		for _, def := range output.QueryDefinitions {
			if aws.ToString(def.Name) == name {
				return &def
			}
		}
		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}
	return nil
}

func TestComposableComplete(t *testing.T, ctx types.TestContext) {
	t.Run("VerifyTerraformOutputs", func(t *testing.T) {
		opts := ctx.TerratestTerraformOptions()
		id := terraform.Output(t, opts, "id")
		name := terraform.Output(t, opts, "name")
		queryString := terraform.Output(t, opts, "query_string")
		logGroupName := terraform.Output(t, opts, "log_group_name")

		assert.NotEmpty(t, id, "query definition id should be set")
		assert.NotEmpty(t, name, "query definition name should be set")
		assert.Contains(t, queryString, "fields @timestamp", "query string should match example")
		assert.NotEmpty(t, logGroupName, "log group name should be set")
	})

	t.Run("VerifyQueryDefinitionViaAPI", func(t *testing.T) {
		opts := ctx.TerratestTerraformOptions()
		queryName := terraform.Output(t, opts, "name")
		region := terraform.Output(t, opts, "region")
		expectedQuery := terraform.Output(t, opts, "query_string")
		expectedLogGroup := terraform.Output(t, opts, "log_group_name")
		expectedID := terraform.Output(t, opts, "id")

		client := getCloudWatchLogsClient(t, region)

		var def *cwltypes.QueryDefinition
		for i := 0; i < 12; i++ {
			def = findQueryDefinitionByName(t, client, queryName)
			if def != nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
		require.NotNil(t, def, "query definition should exist")

		assert.Equal(t, expectedID, aws.ToString(def.QueryDefinitionId), "query definition id should match")
		assert.Equal(t, queryName, aws.ToString(def.Name), "query definition name should match")
		assert.Equal(t, strings.TrimSpace(expectedQuery), strings.TrimSpace(aws.ToString(def.QueryString)), "query string should match")
		assert.Contains(t, def.LogGroupNames, expectedLogGroup, "log group names should include example log group")
	})

	t.Run("StartQueryAndWaitForCompletion", func(t *testing.T) {
		opts := ctx.TerratestTerraformOptions()
		region := terraform.Output(t, opts, "region")
		logGroupName := terraform.Output(t, opts, "log_group_name")
		queryString := terraform.Output(t, opts, "query_string")

		client := getCloudWatchLogsClient(t, region)
		endTime := time.Now().Unix()
		startTime := endTime - 3600

		startOutput, err := client.StartQuery(context.TODO(), &cloudwatchlogs.StartQueryInput{
			LogGroupName: aws.String(logGroupName),
			QueryString:  aws.String(queryString),
			StartTime:    aws.Int64(startTime),
			EndTime:      aws.Int64(endTime),
		})
		require.NoError(t, err, "StartQuery should succeed")
		require.NotEmpty(t, aws.ToString(startOutput.QueryId), "query id should be returned")

		var status cwltypes.QueryStatus
		for i := 0; i < 30; i++ {
			result, err := client.GetQueryResults(context.TODO(), &cloudwatchlogs.GetQueryResultsInput{
				QueryId: startOutput.QueryId,
			})
			require.NoError(t, err, "GetQueryResults should succeed")
			status = result.Status
			if status == cwltypes.QueryStatusComplete || status == cwltypes.QueryStatusFailed || status == cwltypes.QueryStatusCancelled {
				break
			}
			time.Sleep(2 * time.Second)
		}
		assert.Equal(t, cwltypes.QueryStatusComplete, status, "query should complete")
	})
}

func TestComposableCompleteReadOnly(t *testing.T, ctx types.TestContext) {
	t.Run("VerifyTerraformOutputs", func(t *testing.T) {
		opts := ctx.TerratestTerraformOptions()
		id := terraform.Output(t, opts, "id")
		name := terraform.Output(t, opts, "name")

		assert.NotEmpty(t, id, "query definition id should be set")
		assert.NotEmpty(t, name, "query definition name should be set")
	})

	t.Run("VerifyQueryDefinitionExistsViaAPI", func(t *testing.T) {
		opts := ctx.TerratestTerraformOptions()
		queryName := terraform.Output(t, opts, "name")
		region := terraform.Output(t, opts, "region")
		expectedID := terraform.Output(t, opts, "id")

		client := getCloudWatchLogsClient(t, region)

		def := findQueryDefinitionByName(t, client, queryName)
		require.NotNil(t, def, "query definition should exist")
		assert.Equal(t, expectedID, aws.ToString(def.QueryDefinitionId), "query definition id should match")
	})
}
