//go:build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/bitfield/script"
)

func testImpl(impl string) (err error) {
	fmt.Println("Testing", impl)

	// Implementations that don't require a separate service

	switch impl {
	case "badgerdb", "bbolt", "bigcache", "file", "freecache", "gomap", "leveldb", "syncmap", "noop":
		if err = os.Chdir("./" + impl); err != nil {
			return err
		}
		defer os.Chdir("..") // This swallows the error in case there is one, but that's okay as the mage process is exited anyway

		var out string
		out, err = script.Exec("go test -v -race -coverprofile=coverage.txt -covermode=atomic").String()
		fmt.Println(out)
		return err
	}

	// Implementations that require a separate service

	var dockerImage string
	dockerCmd := "docker run -d --rm"
	var setup func() error
	// TODO: Check quoting on Windows
	switch impl {
	case "cockroachdb":
		dockerImage = "cockroachdb/cockroach"
		dockerCmd += ` --name cockroachdb -p 26257:26257 --health-cmd='curl -f http://localhost:8080/health?ready=1' ` + dockerImage + ` start-single-node --insecure`
		setup = func() error {
			var out string
			out, err = script.Exec(`docker exec cockroachdb bash -c './cockroach sql --insecure --execute="create database gokv;"'`).String()
			if err != nil {
				// Print the output here, as it could be more info than what's in err.
				fmt.Println(out)
				return err
			}
			return nil
		}
	case "consul":
		dockerImage = "bitnami/consul"
		dockerCmd += ` --name consul -e CONSUL_LOCAL_CONFIG='{"limits":{"http_max_conns_per_client":1000}}' -p 8500:8500 ` + dockerImage
	case "datastore": // Google Cloud Datastore via "Cloud Datastore Emulator"
		// Using the ":slim" or ":alpine" tag would require the emulator to be installed manually.
		// Both ways seem to be okay for setting the project: `-e CLOUDSDK_CORE_PROJECT=gokv` and CLI parameter `--project=gokv`
		// `--host-port` is required because otherwise the server only listens on localhost IN the container.
		dockerImage = "google/cloud-sdk"
		dockerCmd += ` --name datastore -p 8081:8081 ` + dockerImage + ` gcloud beta emulators datastore start --no-store-on-disk --project=gokv --host-port=0.0.0.0:8081`
	case "dynamodb": // DynamoDB via "DynamoDB local"
		dockerImage = "amazon/dynamodb-local"
		dockerCmd += ` --name dynamodb-local -p 8000:8000 ` + dockerImage
	case "etcd":
		dockerImage = "bitnami/etcd"
		dockerCmd += ` --name etcd -p 2379:2379 --env ALLOW_NONE_AUTHENTICATION=yes --health-cmd='etcdctl endpoint health' ` + dockerImage
	case "hazelcast":
		dockerImage = "hazelcast/hazelcast"
		dockerCmd += ` --name hazelcast -p 5701:5701 --health-cmd='curl -f http://localhost:5701/hazelcast/health/node-state' ` + dockerImage
	case "ignite":
		dockerImage = "apacheignite/ignite"
		dockerCmd += ` --name ignite -p 10800:10800 --health-cmd='${IGNITE_HOME}/bin/control.sh --baseline | grep "Cluster state: active"' ` + dockerImage
	case "memcached":
		dockerImage = "memcached"
		dockerCmd += ` --name memcached -p 11211:11211 ` + dockerImage
	case "mongodb":
		dockerImage = "mongo"
		dockerCmd += ` --name mongodb -p 27017:27017 --health-cmd='echo "db.runCommand({ ping: 1 }).ok" | mongosh localhost:27017/test --quiet' ` + dockerImage
	case "mysql":
		dockerImage = "mysql"
		dockerCmd += ` --name mysql -e MYSQL_ALLOW_EMPTY_PASSWORD=true -p 3306:3306 --health-cmd='mysqladmin ping -h localhost' ` + dockerImage
	case "postgresql":
		dockerImage = "postgres:alpine"
		dockerCmd += ` --name postgres -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=gokv -p 5432:5432 --health-cmd='pg_isready -U postgres' ` + dockerImage
	case "redis":
		dockerImage = "redis"
		dockerCmd += ` --name redis -p 6379:6379 --health-cmd='redis-cli ping' ` + dockerImage
	case "s3": // Amazon S3 via Minio
		dockerImage = "minio/minio"
		dockerCmd += ` --name s3 -e "MINIO_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE" -e "MINIO_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" -p 9000:9000 --health-cmd='mc ready local' ` + dockerImage + ` server /data`
	case "tablestorage": // Tablestorage via Azurite
		// In the past there was this problem: https://github.com/Azure/Azurite/issues/121
		// With this Docker image:
		// docker run -d --rm --name azurite -e executable=table -p 10002:10002 arafato/azurite
		// Now with the official image it still doesn't work. // TODO: Investigate / create GitHub issue.
		// docker run -d --rm --name azurite -p 10002:10002 mcr.microsoft.com/azure-storage/azurite azurite-table
	case "tablestore":
		// Currently no emulator exists for Alibaba Cloud Table Store.
	case "zookeeper":
		dockerImage = "zookeeper"
		dockerCmd += ` --name zookeeper -p 2181:2181 -e ZOO_4LW_COMMANDS_WHITELIST=ruok --health-cmd='echo ruok | timeout 2 nc -w 2 localhost 2181 | grep imok' ` + dockerImage
	default:
		return errors.New("unknown `gokv.Store` implementation")
	}

	// TODO: until docker images for windows appear, skip those test for windows
	if dockerImage != "" && runtime.GOOS == "windows" {
		return nil
	}

	// For some implementations there's no way to test with a Docker container yet.
	// For them we skip the Docker stuff but still execute the tests, which can skip on connection error and we can see the skips in the test results.
	if dockerImage != "" {
		// Pull Docker image
		var out string
		out, err = script.Exec("docker pull " + dockerImage + ":latest").String()
		if err != nil {
			// Depending on the error, printing the output could be interesting, as it could be more info than what's in err.
			fmt.Println(out)
			return err
		}
		// Start Docker container
		out, err = script.Exec(dockerCmd).String()
		if err != nil {
			fmt.Println(out)
			return err
		}
		// Thanks to separate pull and run, the output of the run is only the container ID
		// instead of including pull progress.
		containerID := strings.ReplaceAll(out, "\n", "")
		defer func() {
			out, err2 := script.Exec("docker stop " + containerID).String()
			if err2 != nil {
				// Set err for returning, but only if it's not set yet.
				// ⚠️ Make sure all below errors are set to `err` with `err = ` and not shadowed with `:=`.
				if err == nil {
					err = err2
				}
				// Depending on the error, printing the output could be interesting, as it could be more info than what's in err.
				fmt.Println(out)
			}
		}()

		// Wait for container to be started
		if strings.Contains(dockerCmd, "--health-cmd") {
			for i := 0; i < 10; i++ {
				out, err = script.Exec("docker inspect --format='{{.State.Health.Status}}' " + containerID).String()
				if err != nil {
					fmt.Println(out)
					return err
				}
				out = strings.ReplaceAll(out, "\n", "")
				if out == "healthy" {
					break
				}
				fmt.Printf("Waiting for container to be healthy... (%d/10)\n", i+1)
				time.Sleep(time.Second)
			}
		} else {
			fmt.Println("Waiting 10 seconds for container due to missing health check...")
			time.Sleep(10 * time.Second)
		}

		if setup != nil {
			err = setup()
			if err != nil {
				return err
			}
		}
	}

	err = os.Chdir(impl)
	if err != nil {
		return err
	}
	defer os.Chdir("..") // This swallows the error in case there is one, but that's okay as the mage process is exited anyway

	var out string
	out, err = script.Exec("go test -v -race -coverprofile=coverage.txt -covermode=atomic").String()
	fmt.Println(out)

	// If err is nil, the above deferred functions might set it
	return err
}
