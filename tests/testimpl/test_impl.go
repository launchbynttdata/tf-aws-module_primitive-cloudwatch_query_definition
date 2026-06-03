package testimpl

import (
	"context"
	"fmt"
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

const expectedExampleQueryString = "fields @timestamp, @message\n| sort @timestamp desc\n| limit 25"

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

func describeLogGroupWithRetry(t *testing.T, client *cloudwatchlogs.Client, logGroupName string) *cwltypes.LogGroup {
	t.Helper()
	var lastErr error
	for attempt := 1; attempt <= 12; attempt++ {
		output, err := client.DescribeLogGroups(context.TODO(), &cloudwatchlogs.DescribeLogGroupsInput{
			LogGroupNamePrefix: aws.String(logGroupName),
		})
		if err != nil {
			lastErr = err
			time.Sleep(5 * time.Second)
			continue
		}
		for i := range output.LogGroups {
			if aws.ToString(output.LogGroups[i].LogGroupName) == logGroupName {
				return &output.LogGroups[i]
			}
		}
		lastErr = fmt.Errorf("log group %s not found yet", logGroupName)
		time.Sleep(5 * time.Second)
	}
	require.Failf(t, "unable to describe log group", "%s after retries: %v", logGroupName, lastErr)
	return nil
}

func assertLogGroupKMSEncryption(t *testing.T, client *cloudwatchlogs.Client, logGroupName, expectedKMSArn string) {
	t.Helper()
	logGroup := describeLogGroupWithRetry(t, client, logGroupName)
	require.NotNil(t, logGroup.KmsKeyId, "log group should have a KMS key attached")
	assert.Equal(t, expectedKMSArn, aws.ToString(logGroup.KmsKeyId), "KMS key ARN should match")
}

func waitForQueryDefinition(t *testing.T, client *cloudwatchlogs.Client, name string) *cwltypes.QueryDefinition {
	t.Helper()
	for i := 0; i < 12; i++ {
		if def := findQueryDefinitionByName(t, client, name); def != nil {
			return def
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}

func assertQueryDefinitionConfig(t *testing.T, def *cwltypes.QueryDefinition, opts *terraform.Options, queryName string) {
	t.Helper()
	expectedID := terraform.Output(t, opts, "id")
	expectedQuery := terraform.Output(t, opts, "query_string")
	expectedLogGroup := terraform.Output(t, opts, "log_group_name")

	require.NotNil(t, def, "query definition should exist")
	assert.Equal(t, expectedID, aws.ToString(def.QueryDefinitionId), "query definition id should match")
	assert.Equal(t, queryName, aws.ToString(def.Name), "query definition name should match")
	assert.Equal(t, strings.TrimSpace(expectedQuery), strings.TrimSpace(aws.ToString(def.QueryString)), "query string should match")
	assert.Contains(t, def.LogGroupNames, expectedLogGroup, "log group names should include example log group")
}

func TestComposableComplete(t *testing.T, ctx types.TestContext) {
	t.Run("VerifyTerraformOutputs", func(t *testing.T) {
		opts := ctx.TerratestTerraformOptions()
		id := terraform.Output(t, opts, "id")
		name := terraform.Output(t, opts, "name")
		queryString := terraform.Output(t, opts, "query_string")
		logGroupName := terraform.Output(t, opts, "log_group_name")

		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, id, "query definition id should be a UUID")
		assert.Equal(t, name, terraform.Output(t, opts, "name"), "name output should be stable")
		assert.Equal(t, expectedExampleQueryString, strings.TrimSpace(queryString), "query string should match example")
		assert.Regexp(t, `^/aws/example/`, logGroupName, "log group name should use example prefix")
	})

	t.Run("VerifyLogGroupKMSEncryption", func(t *testing.T) {
		opts := ctx.TerratestTerraformOptions()
		logGroupName := terraform.Output(t, opts, "log_group_name")
		kmsKeyARN := terraform.Output(t, opts, "kms_key_arn")
		region := terraform.Output(t, opts, "region")

		client := getCloudWatchLogsClient(t, region)
		assertLogGroupKMSEncryption(t, client, logGroupName, kmsKeyARN)
	})

	t.Run("VerifyQueryDefinitionViaAPI", func(t *testing.T) {
		opts := ctx.TerratestTerraformOptions()
		queryName := terraform.Output(t, opts, "name")
		region := terraform.Output(t, opts, "region")

		client := getCloudWatchLogsClient(t, region)
		def := waitForQueryDefinition(t, client, queryName)
		assertQueryDefinitionConfig(t, def, opts, queryName)
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
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, aws.ToString(startOutput.QueryId), "query id should be a UUID")

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

		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, id, "query definition id should be a UUID")
		assert.Equal(t, name, terraform.Output(t, opts, "name"), "name output should be stable")
	})

	t.Run("VerifyLogGroupKMSEncryption", func(t *testing.T) {
		opts := ctx.TerratestTerraformOptions()
		logGroupName := terraform.Output(t, opts, "log_group_name")
		kmsKeyARN := terraform.Output(t, opts, "kms_key_arn")
		region := terraform.Output(t, opts, "region")

		client := getCloudWatchLogsClient(t, region)
		assertLogGroupKMSEncryption(t, client, logGroupName, kmsKeyARN)
	})

	t.Run("VerifyQueryDefinitionViaAPI", func(t *testing.T) {
		opts := ctx.TerratestTerraformOptions()
		queryName := terraform.Output(t, opts, "name")
		region := terraform.Output(t, opts, "region")

		client := getCloudWatchLogsClient(t, region)
		def := waitForQueryDefinition(t, client, queryName)
		assertQueryDefinitionConfig(t, def, opts, queryName)
	})
}
