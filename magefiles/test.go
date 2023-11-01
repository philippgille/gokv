//go:build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bitfield/script"
)

func testImpl(impl string) error {
	fmt.Println("Testing", impl)

	// Implementations that don't require a separate service

	switch impl {
	case "badgerdb", "bbolt", "bigcache", "file", "freecache", "gomap", "leveldb", "syncmap", "noop":
		if err := os.Chdir("./" + impl); err != nil {
			return err
		}
		defer os.Chdir("..") // This swallows the error in case there is one, but that's okay as the mage process is exited anyway

		out, err := script.Exec("go test -v -race -coverprofile=coverage.txt -covermode=atomic").String()
		// In case of a test error, just returning err wouldn't print the go test output details, so we need to print out as well.
		fmt.Println(out)
		return err
	}

	// Implementations that require a separate service

	var err error
	var dockerCmd string
	var setup func() error
	// TODO: Check quoting on Windows
	switch impl {
	case "cockroachdb":
		dockerCmd = `docker run -d --rm --name cockroachdb -p 26257:26257 --health-cmd='curl -f http://localhost:8080/health?ready=1' cockroachdb/cockroach start-single-node --insecure`
		setup = func() error {
			out, err := script.Exec(`docker exec cockroachdb bash -c './cockroach sql --insecure --execute="create database gokv;"'`).String()
			if err != nil {
				// Print the output here, as it could be more info than what's in err.
				fmt.Println(out)
				return err
			}
			return nil
		}
	case "consul":
		dockerCmd = `docker run -d --rm --name consul -e CONSUL_LOCAL_CONFIG='{"limits":{"http_max_conns_per_client":1000}}' -p 8500:8500 bitnami/consul`
	case "datastore": // Google Cloud Datastore via "Cloud Datastore Emulator"
		// Using the ":slim" or ":alpine" tag would require the emulator to be installed manually.
		// Both ways seem to be okay for setting the project: `-e CLOUDSDK_CORE_PROJECT=gokv` and CLI parameter `--project=gokv`
		// `--host-port` is required because otherwise the server only listens on localhost IN the container.
		dockerCmd = `docker run -d --rm --name datastore -p 8081:8081 google/cloud-sdk gcloud beta emulators datastore start --no-store-on-disk --project=gokv --host-port=0.0.0.0:8081`
	case "dynamodb": // DynamoDB via "DynamoDB local"
		dockerCmd = `docker run -d --rm --name dynamodb-local -p 8000:8000 amazon/dynamodb-local`
	case "etcd":
		dockerCmd = `docker run -d --rm --name etcd -p 2379:2379 --env ALLOW_NONE_AUTHENTICATION=yes --health-cmd='etcdctl endpoint health' bitnami/etcd`
	case "hazelcast":
		dockerCmd = `docker run -d --rm --name hazelcast -p 5701:5701 --health-cmd='curl -f http://localhost:5701/hazelcast/health/node-state' hazelcast/hazelcast`
	case "ignite":
		dockerCmd = `docker run -d --rm --name ignite -p 10800:10800 --health-cmd='${IGNITE_HOME}/bin/control.sh --baseline | grep "Cluster state: active"' apacheignite/ignite`
	case "memcached":
		dockerCmd = `docker run -d --rm --name memcached -p 11211:11211 memcached`
	case "mongodb":
		dockerCmd = `docker run -d --rm --name mongodb -p 27017:27017 --health-cmd='echo "db.runCommand({ ping: 1 }).ok" | mongosh localhost:27017/test --quiet' mongo`
	case "mysql":
		dockerCmd = `docker run -d --rm --name mysql -e MYSQL_ALLOW_EMPTY_PASSWORD=true -p 3306:3306 --health-cmd='mysqladmin ping -h localhost' mysql`
	case "postgresql":
		dockerCmd = `docker run -d --rm --name postgres -e POSTGRES_PASSWORD=secret -e POSTGRES_DB=gokv -p 5432:5432 --health-cmd='pg_isready -U postgres' postgres:alpine`
	case "redis":
		dockerCmd = `docker run -d --rm --name redis -p 6379:6379 --health-cmd='redis-cli ping' redis`
	case "s3": // Amazon S3 via Minio
		dockerCmd = `docker run -d --rm --name s3 -e "MINIO_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE" -e "MINIO_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" -p 9000:9000 --health-cmd='mc ready local' minio/minio server /data`
	case "tablestorage": // Tablestorage via Azurite
		// In the past there was this problem: https://github.com/Azure/Azurite/issues/121
		// With this Docker image:
		//docker run -d --rm --name azurite -e executable=table -p 10002:10002 arafato/azurite
		// Now with the official image it still doesn't work. // TODO: Investigate / create GitHub issue.
		//docker run -d --rm --name azurite -p 10002:10002 mcr.microsoft.com/azure-storage/azurite azurite-table
	case "tablestore":
		// Currently no emulator exists for Alibaba Cloud Table Store.
	case "zookeeper":
		dockerCmd = `docker run -d --rm --name zookeeper -p 2181:2181 -e ZOO_4LW_COMMANDS_WHITELIST=ruok --health-cmd='echo ruok | timeout 2 nc -w 2 localhost 2181 | grep imok' zookeeper`
	default:
		return errors.New("unknown `gokv.Store` implementation")
	}

	// For some implementations there's no way to test with a Docker container yet.
	// For them we skip the Docker stuff but still execute the tests, which can skip on connection error and we can see the skips in the test results.
	if dockerCmd != "" {
		// Start Docker container
		var out string
		out, err = script.Exec(dockerCmd).String()
		if err != nil {
			// Depending on the error, printing the output could be interesting, as it could be more info than what's in err.
			fmt.Println(out)
			return err
		}
		// In the success case, if the image was pulled as part of `docker run`, the output is not only the container ID, but also the pull progress.
		outLines := strings.Split(out, "\n")
		containerID := outLines[len(outLines)-1]
		// Docker output could end with a newline, in which case we use the previous line
		if containerID == "" {
			containerID = outLines[len(outLines)-2]
		}
		defer func() {
			out, err2 := script.Exec("docker stop " + containerID).String()
			if err2 != nil {
				// Set err for returning, but only if it's not set yet
				if err == nil {
					err = err2
				}
				// Set the err var from outer scope, so it's the one returned from the outer function.
				// Depending on the error, printing the output could be interesting, as it could be more info than what's in err.
				fmt.Println(out)
			}
		}()

		// Wait for container to be started
		// TODO: Use a proper health/startup check for this, as many services are ready much quicker than 10s.
		time.Sleep(10 * time.Second)
		out, _ = script.Exec("docker inspect --format='{{.State.Health.Status}}' " + containerID).String()
		fmt.Println(out)

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

	out, err := script.Exec("go test -v -race -coverprofile=coverage.txt -covermode=atomic").String()
	// In case of a test error, just returning err wouldn't print the go test output details, so we need to print out as well.
	fmt.Println(out)

	// If err is nil, the above deferred functions might set it
	return err
}
