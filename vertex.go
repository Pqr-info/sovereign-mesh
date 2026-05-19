package sovereign

import (
	"context"
	"fmt"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/structpb"
)

// OptimizePrompt triggers a Vertex AI prompt optimization job.
func (c *Controller) OptimizePrompt(ctx context.Context, task OptimizationTask) (string, error) {
	endpoint := fmt.Sprintf("%s-aiplatform.googleapis.com:443", c.location)
	client, err := aiplatform.NewPipelineClient(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		return "", err
	}
	defer client.Close()

	params, _ := structpb.NewValue(map[string]interface{}{
		"prompt":      task.Prompt,
		"method":      task.Method,
		"dataset_uri": task.ExamplePath,
	})

	req := &aiplatformpb.CreatePipelineJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", c.projectID, c.location),
		PipelineJob: &aiplatformpb.PipelineJob{
			DisplayName: "Sovereign-Opt-" + task.ID,
			TemplateUri: "https://us-kfp.pkg.dev/ml-pipeline/google-cloud-pipeline-components/vertex-ai-prompt-optimizer/v1",
			RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{
				ParameterValues: map[string]*structpb.Value{
					"optimization_spec": params,
				},
			},
		},
	}

	job, err := client.CreatePipelineJob(ctx, req)
	if err != nil {
		return "", err
	}

	c.syncLock.Lock()
	c.metrics["vertex/jobs/created"]++
	c.syncLock.Unlock()

	return job.GetName(), nil
}