package bitbucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeploymentEndpoints(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/2.0/repositories/acme/widgets/deployments/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{dep-1}","state":{"name":"SUCCESSFUL"},"environment":{"uuid":"{env-1}","name":"Test","slug":"test"},"release":{"name":"main"}}]}`))
	})
	mux.HandleFunc("/2.0/repositories/acme/widgets/deployments/%7Bdep-1%7D", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"uuid":"{dep-1}","state":{"name":"SUCCESSFUL"},"environment":{"uuid":"{env-1}","name":"Test","slug":"test"},"release":{"name":"main"}}`))
	})
	mux.HandleFunc("/2.0/repositories/acme/widgets/environments/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{env-1}","name":"Test","slug":"test","category":{"name":"Test"},"lock":{"name":"OPEN"}}]}`))
	})
	mux.HandleFunc("/2.0/repositories/acme/widgets/environments/%7Benv-1%7D", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"uuid":"{env-1}","name":"Test","slug":"test","category":{"name":"Test"},"lock":{"name":"OPEN"}}`))
	})
	mux.HandleFunc("/2.0/repositories/acme/widgets/deployments_config/environments/%7Benv-1%7D/variables/%7Bvar-1%7D", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"uuid":"{var-1}","key":"APP_ENV","secured":true}`))
		case http.MethodPut:
			_, _ = w.Write([]byte(`{"uuid":"{var-1}","key":"APP_ENV","secured":true}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("/2.0/repositories/acme/widgets/deployments_config/environments/%7Benv-1%7D/variables", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			_, _ = w.Write([]byte(`{"uuid":"{var-2}","key":"NEW_KEY","secured":false}`))
			return
		}
		_, _ = w.Write([]byte(`{"values":[{"uuid":"{var-1}","key":"APP_ENV","secured":true}]}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("BB_API_BASE_URL", server.URL+"/2.0")
	client := repositoryTestClient(t)

	deployments, err := client.ListDeployments(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListDeployments returned error: %v", err)
	}
	if len(deployments) != 1 || deployments[0].UUID != "{dep-1}" {
		t.Fatalf("unexpected deployments %+v", deployments)
	}

	deployment, err := client.GetDeployment(context.Background(), "acme", "widgets", "dep-1")
	if err != nil {
		t.Fatalf("GetDeployment returned error: %v", err)
	}
	if deployment.UUID != "{dep-1}" || deployment.Environment.Slug != "test" {
		t.Fatalf("unexpected deployment %+v", deployment)
	}

	environments, err := client.ListDeploymentEnvironments(context.Background(), "acme", "widgets", 10)
	if err != nil {
		t.Fatalf("ListDeploymentEnvironments returned error: %v", err)
	}
	if len(environments) != 1 || environments[0].Slug != "test" {
		t.Fatalf("unexpected environments %+v", environments)
	}

	environment, err := client.GetDeploymentEnvironment(context.Background(), "acme", "widgets", "env-1")
	if err != nil {
		t.Fatalf("GetDeploymentEnvironment returned error: %v", err)
	}
	if environment.UUID != "{env-1}" || environment.Name != "Test" {
		t.Fatalf("unexpected environment %+v", environment)
	}

	variables, err := client.ListDeploymentVariables(context.Background(), "acme", "widgets", "env-1", ListDeploymentVariablesOptions{Limit: 10})
	if err != nil {
		t.Fatalf("ListDeploymentVariables returned error: %v", err)
	}
	if len(variables) != 1 || variables[0].Key != "APP_ENV" {
		t.Fatalf("unexpected deployment variables %+v", variables)
	}

	variable, err := client.GetDeploymentVariable(context.Background(), "acme", "widgets", "env-1", "var-1")
	if err != nil {
		t.Fatalf("GetDeploymentVariable returned error: %v", err)
	}
	if variable.UUID != "{var-1}" || variable.Key != "APP_ENV" {
		t.Fatalf("unexpected deployment variable %+v", variable)
	}

	created, err := client.CreateDeploymentVariable(context.Background(), "acme", "widgets", "env-1", DeploymentVariable{Key: "NEW_KEY", Value: "hello"})
	if err != nil {
		t.Fatalf("CreateDeploymentVariable returned error: %v", err)
	}
	if created.UUID != "{var-2}" || created.Key != "NEW_KEY" {
		t.Fatalf("unexpected created deployment variable %+v", created)
	}

	updated, err := client.UpdateDeploymentVariable(context.Background(), "acme", "widgets", "env-1", "var-1", DeploymentVariable{Key: "APP_ENV", Value: "updated", Secured: true})
	if err != nil {
		t.Fatalf("UpdateDeploymentVariable returned error: %v", err)
	}
	if updated.UUID != "{var-1}" || updated.Key != "APP_ENV" || !updated.Secured {
		t.Fatalf("unexpected updated deployment variable %+v", updated)
	}

	if err := client.DeleteDeploymentVariable(context.Background(), "acme", "widgets", "env-1", "var-1"); err != nil {
		t.Fatalf("DeleteDeploymentVariable returned error: %v", err)
	}
}
