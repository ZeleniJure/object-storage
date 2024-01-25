package storagebackend

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
)

type S3Storage struct {
	clients []*minio.Client
}

type instance struct {
	name      string
	address   string
	accessKey string
	keyId     string
	ctx       context.Context
	ctxCancel context.CancelFunc
	minio     *minio.Client
}

// TODO we have config...we could read it from there
const EnvAccessKey = "MINIO_ACCESS_KEY"
const EnvKeyId = "MINIO_SECRET_KEY"
const bucket = "amazing-bucket"
const minioPort = ":9000"

var clients []*minio.Client
var cm sync.Mutex

func New() *S3Storage {
	if len(clients) <= 0 {
		cm.Lock()
		defer cm.Unlock()
		if len(clients) <= 0 {
			instances := findMinio()
			connectInstances(instances)
			clients = prepareBuckets(instances)
		}
		if len(clients) <= 0 {
			log.Panic().Msg("No backend storage detected, exiting...")
		}
	}

	return &S3Storage{clients}
}

func (s *S3Storage) Put(name string, content io.Reader) error {
	client, err := s.getInstance(name)
	if err != nil {
		return err
	}
	_, err = client.PutObject(
		context.Background(),
		bucket,
		name,
		content,
		-1,
		minio.PutObjectOptions{},
	)
	return err
}

func (s *S3Storage) Get(ctx context.Context, name string) (*bytes.Buffer, error) {
	client, err := s.getInstance(name)
	if err != nil {
		return nil, err
	}
	object, err := client.GetObject(
		ctx,
		bucket,
		name,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, err
	}
	content := &bytes.Buffer{}
	_, err = io.Copy(content, object)
	return content, err
}

func (s *S3Storage) getInstance(name string) (*minio.Client, error) {
	if len(s.clients) <= 0 {
		return nil, errors.New("No backend storage detected.")
	}
	var sum int
	for _, letter := range name {
		sum += int(letter)
	}
	id := sum % len(s.clients)
	client := s.clients[id]
	log.Info().Str("host", client.EndpointURL().Host).Int("id", id).Msg("Backend storage selected")
	return client, nil
}

func connectInstances(instances []*instance) {
	for i := 0; i < len(instances); i++ {
		minioClient, err := minio.New(instances[i].address, &minio.Options{
			Creds:  credentials.NewStaticV4(instances[i].accessKey, instances[i].keyId, ""),
			Secure: false,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Could not create minio client")
		}

		instances[i].minio = minioClient
	}
}

func prepareBuckets(instances []*instance) []*minio.Client {
	var clients []*minio.Client
	// TODO would be nice if clients get sorted based on instance name or something
	for _, instance := range instances {
		client := instance.minio
		if client == nil {
			instance.ctxCancel()
			continue
		}
		exists, _ := client.BucketExists(instance.ctx, bucket)
		if exists {
			clients = append(clients, client)
			continue
		}
		err := client.MakeBucket(instance.ctx, bucket, minio.MakeBucketOptions{ObjectLocking: true})
		if err != nil {
			instance.ctxCancel()
			log.Err(err).Stack().Str("name", instance.name).Str("host", instance.address).Msg("Can't create bucket")
			continue
		}
		clients = append(clients, client)
	}
	return clients
}

func findMinio() []*instance {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Panic().Err(err).Msg("Could not connect to docker deamon")
	}
	defer apiClient.Close()

	ctxTimeout, _ := context.WithTimeout(context.Background(), time.Second*10)
	containers, err := apiClient.ContainerList(ctxTimeout, container.ListOptions{})
	if err != nil {
		log.Panic().Err(err).Msg("Problem trying to list containers")
	}

	var instances []*instance
	for _, ctr := range containers {

		if ctr.State != "running" {
			// TODO: would be nice to log these in real life, so that
			//       users can debug why containers aren't detected
			continue
		}
		name := contains(ctr.Names, "amazin-object-storage-node")
		if name == "" {
			continue
		}
		networks := maps.Keys(ctr.NetworkSettings.Networks)
		network := contains(networks, "amazin-object-storage")
		if network == "" {
			continue
		}
		// Does IP address really exist? I trust docker API to actually have all entries...hopefully no null pointer exceptions will follow!
		address := ctr.NetworkSettings.Networks[network].IPAddress
		log.Info().Str("name", name).Str("address", address).Msg("Found container, extracting minio credentials...")

		td := time.Duration(viper.GetInt("minio.searchTimeout")) * time.Second
		containerTimeout, cancel := context.WithTimeout(context.Background(), td)
		inspect, err := apiClient.ContainerInspect(containerTimeout, ctr.ID)
		if err != nil {
			cancel()
			continue
		}
		accessKey := contains(inspect.Config.Env, EnvAccessKey)
		secretKey := contains(inspect.Config.Env, EnvKeyId)
		if accessKey == "" || secretKey == "" {
			cancel()
			continue
		}
		instances = append(
			instances,
			&instance{
				name,
				address + minioPort,
				accessKey[len(EnvAccessKey)+1:],
				secretKey[len(EnvKeyId)+1:],
				containerTimeout,
				cancel,
				nil,
			},
		)
	}
	return instances
}

func contains(list []string, find string) string {
	for _, item := range list {
		if strings.Contains(item, find) {
			return item
		}
	}
	return ""
}
