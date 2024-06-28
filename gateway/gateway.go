// Package gateway provides a wrapper around the Invopop gateway service used
// to respond to incoming messages to process tasks.
//
// This package is only meant to be used by applications that will receive and
// process tasks via NATS, hence why it is independent from the main Invopop
// package.
package gateway

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gabriel-vasile/mimetype"
	"github.com/invopop/configure/pkg/natsconf"
	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/encoders/protobuf"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

const (
	defaultWorkerCount = 8
)

// Client wraps around the functionality provided by the
// the gateway service, accessed via NATS.
type Client struct {
	name              string // service name
	nc                *nats.EncodedConn
	wg                sync.WaitGroup
	th                TaskHandler
	incoming          chan *nats.Msg
	sub               *nats.Subscription
	siloPublicBaseURL string
	workerCount       int
}

// New instantiates a new gateway client service using the provided config.
func New(conf Configuration) *Client {
	gw := new(Client)
	gconf := conf.config()

	gw.workerCount = gconf.WorkerCount
	if gw.workerCount == 0 {
		gw.workerCount = defaultWorkerCount
	}
	gw.name = gconf.Name
	gw.nc = prepareNATSClient(gconf.NATS, gconf.Name)
	gw.incoming = make(chan *nats.Msg)

	if gconf.Silo != nil {
		gw.siloPublicBaseURL = gconf.Silo.PublicBaseURL
	}

	return gw
}

// NATS provides the NATS Encoded Connection so that it can be used
// for other tasks if needed.
func (gw *Client) NATS() *nats.EncodedConn {
	return gw.nc
}

// Subscribe indicates which method should be called when
// messages are received from the gateway service.
func (gw *Client) Subscribe(th TaskHandler) {
	gw.th = th
}

// Poke sends a message to the gateway indicating that we've received an
// external prompt, like a webhook, and the original task should be re-sent.
func (gw *Client) Poke(ctx context.Context, req *TaskPoke) error {
	res := new(TaskPokeResponse)
	if err := gw.nc.RequestWithContext(ctx, SubjectTasksPoke, req, res); err != nil {
		return err
	}
	if res.Err != nil {
		return res.Err
	}
	// PokeTaskResponse is empty if successful
	return nil
}

// CreateFile allows us to build a file place holder and upload the data afterwards
// by posting to the URL provided.
func (gw *Client) CreateFile(ctx context.Context, req *CreateFile) (*File, error) {
	res := new(FileResponse)
	if err := gw.nc.RequestWithContext(ctx, SubjectFilesCreate, req, res); err != nil {
		return nil, err
	}
	if res.Err != nil {
		return nil, res.Err
	}
	return res.File, nil
}

// CreateAndUploadFile makes it easier to upload a file using basic details such
// as the ID, Job, Envelope, Name, and Description, and automatically add the
// SHA256, MIME, and Size attributes based on the provided data.
// The tradeoff here as opposed to two separate calls is that the data is kept
// in memory and not sent through a buffer.
func (gw *Client) CreateAndUploadFile(ctx context.Context, req *CreateFile, data []byte) (*File, error) {
	gw.prepareCreateFileFromData(req, data)

	f, err := gw.CreateFile(ctx, req)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(data)
	if err := gw.UploadFile(ctx, f, r); err != nil {
		return nil, err
	}

	return f, nil
}

// UploadFile performs an HTTP PUT action to send the data to the silo. The URL comes
// from the file object. Provided the SHA256 data matches, the file uploaded will
// function as expected.
func (gw *Client) UploadFile(ctx context.Context, f *File, data io.Reader) error {
	url, err := gw.fileUploadURL(f)
	if err != nil {
		return fmt.Errorf("upload url: %w", err)
	}
	req, err := http.NewRequest(http.MethodPut, url, data)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", f.Mime)
	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("upload error, status: %s", res.Status)
	}
	return nil
}

// FetchFile performs an HTTP GET action to retrieve a file's data from the
// silo. The URL comes from the file object.
func (gw *Client) FetchFile(ctx context.Context, f *File) ([]byte, error) {
	url, err := gw.fileUploadURL(f)
	if err != nil {
		return nil, fmt.Errorf("upload url: %w", err)
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	req = req.WithContext(ctx)
	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("fetch error, status: %s", res.Status)
	}
	return io.ReadAll(res.Body)
}

func (gw *Client) prepareCreateFileFromData(req *CreateFile, data []byte) {
	req.Size = int32(len(data))
	mt := mimetype.Detect(data)
	req.Mime = mt.String()
	sum := sha256.Sum256(data)
	req.Sha256 = hex.EncodeToString(sum[:])
}

// fileUploadURL determines the URL used to upload to. We don't use the URL
// provided by the silo service as that is only reliable for exterior public
// use.
func (gw *Client) fileUploadURL(f *File) (string, error) {
	if gw.siloPublicBaseURL == "" {
		return "", errors.New("missing silo public base url")
	}
	return fmt.Sprintf("%s/%s/%s/%s?h=%s",
		gw.siloPublicBaseURL,
		f.SiloEntryId,
		f.Id,
		f.Name,
		f.Hash,
	), nil
}

// Start begins the gateway service and starts listening for incoming tasks.
func (gw *Client) Start(ctx context.Context) error {
	if gw.th == nil {
		return errors.New("task handler required")
	}
	if gw.nc == nil {
		return errors.New("nats connection required")
	}
	if err := gw.subscribeIncomingTasks(); err != nil {
		return fmt.Errorf("subscribing for tasks: %w", err)
	}
	for i := 0; i < gw.workerCount; i++ {
		go gw.startTaskWorker(ctx)
	}
	return nil
}

// Stop is used to gracefully drain all requests and wait for them to complete.
func (gw *Client) Stop() {
	if gw.sub != nil {
		gw.sub.Unsubscribe() // nolint:errcheck
		gw.sub.Drain()       // nolint:errcheck
	}
	close(gw.incoming) // this stops workers from receiving more
	gw.wg.Wait()
}

func (gw *Client) subscribeIncomingTasks() error {
	subj := fmt.Sprintf(SubjectTaskFmt, gw.name)
	queue := fmt.Sprintf(QueueNameTaskFmt, gw.name)
	var err error
	gw.sub, err = gw.nc.Conn.QueueSubscribeSyncWithChan(subj, queue, gw.incoming)
	if err != nil {
		return fmt.Errorf("error subscribing to queue: %w", err)
	}
	return nil
}

func (gw *Client) startTaskWorker(ctx context.Context) {
	for m := range gw.incoming {
		gw.processTask(ctx, m)
	}
}

func (gw *Client) processTask(ctx context.Context, m *nats.Msg) {
	gw.wg.Add(1)
	defer gw.wg.Done()

	// Handling the incoming data
	t := new(Task)
	var res *TaskResult
	if err := proto.Unmarshal(m.Data, t); err != nil {
		res = TaskError(fmt.Errorf("parsing incoming task: %w", err))
	} else {
		res = gw.th(ctx, t)
		if res == nil {
			// assume the response is okay if no content
			res = TaskOK()
		}
	}

	// Send the reply back
	data, err := proto.Marshal(res)
	if err != nil {
		log.Error().Str("task_id", t.Id).Err(err).Msg("unable to marshal task response, dropping")
	}
	if err := gw.nc.Conn.Publish(m.Reply, data); err != nil {
		log.Error().Str("task_id", t.Id).Err(err).Msg("unable to publish response")
	}
}

func prepareNATSClient(conf *natsconf.Config, name string) *nats.EncodedConn {
	// prepare base options
	opts, err := conf.Options()
	if err != nil {
		log.Fatal().Err(err).Msg("preparing nats options")
	}

	opts = append(opts, nats.Name(name))

	// Add our own connection logging stuff
	opts = append(opts, nats.ConnectHandler(func(nc *nats.Conn) {
		log.Info().Str("url", conf.URL).Msg("nats connected")
	}))
	opts = append(opts, nats.DisconnectHandler(func(nc *nats.Conn) {
		log.Warn().Str("url", conf.URL).Msg("nats disconnected")
	}))

	// Create the encoded connection
	nc, err := nats.Connect(conf.URL, opts...)
	if err != nil {
		log.Fatal().Err(err).Str("url", conf.URL).Msg("failed to connect to nats")
	}
	enc, err := nats.NewEncodedConn(nc, protobuf.PROTOBUF_ENCODER)
	if err != nil {
		log.Fatal().Err(err).Str("url", conf.URL).Msg("failed to prepare nats encoded connection")
	}

	return enc
}
