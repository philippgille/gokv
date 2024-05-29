//go:build integration

package couchbase_test

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	couchbase_container "github.com/testcontainers/testcontainers-go/modules/couchbase"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	cleanUp, err := startCouchbaseContainer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to start couchbase container: %v", err)

		os.Exit(1)
	}

	exitCode := m.Run()

	err = cleanUp(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to terminate couchbase container: %v", err)
	}

	os.Exit(exitCode)
}

func startCouchbaseContainer(ctx context.Context) (cleanup func(context.Context) error, err error) {
	bucketName := cmp.Or(os.Getenv(cbBucketNameEnvVar), "testBucket")
	username := cmp.Or(os.Getenv(cbUsernameEnvVar), "username")
	password := cmp.Or(os.Getenv(cbPasswordEnvVar), "password")

	bucket := couchbase_container.NewBucket(bucketName).WithQuota(100).
		WithReplicas(0).
		WithFlushEnabled(false).
		WithPrimaryIndex(true)

	couchbaseContainer, err := couchbase_container.RunContainer(ctx,
		testcontainers.WithImage("couchbase:community-7.1.1"),
		couchbase_container.WithAdminCredentials(username, password),
		couchbase_container.WithBuckets(bucket),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	connStr, err := couchbaseContainer.ConnectionString(ctx)
	if err != nil {
		defer couchbaseContainer.Terminate(ctx)

		return nil, fmt.Errorf("failed to fetch couchbase connection string from container: %w", err)
	}

	os.Setenv(cbConnectionStringEnvVar, connStr)
	os.Setenv(cbBucketNameEnvVar, bucketName)
	os.Setenv(cbUsernameEnvVar, username)
	os.Setenv(cbPasswordEnvVar, password)

	return func(ctx context.Context) error {
		return couchbaseContainer.Terminate(ctx)
	}, nil
}
