package pdkafka

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// -----------------------------------------------------------------------------
// Client + Provider (Sarama)
// -----------------------------------------------------------------------------

type Client struct {
	cfg   *Config
	scfg  *sarama.Config
	cli   sarama.Client
	admin sarama.ClusterAdmin
	sp    sarama.SyncProducer
	ap    sarama.AsyncProducer // used only for transactions
}

func CreateClientFromConfig(cfg *Config) (*Client, error) {
	scfg, err := buildSaramaConfig(cfg)
	if err != nil {
		return nil, err
	}
	cli, err := sarama.NewClient(cfg.Brokers, scfg)
	if err != nil {
		return nil, err
	}
	adm, err := sarama.NewClusterAdminFromClient(cli)
	if err != nil {
		_ = cli.Close()
		return nil, err
	}
	return &Client{cfg: cfg, scfg: scfg, cli: cli, admin: adm}, nil
}

func (c *Client) Close() error {
	var err error
	if c.sp != nil {
		err = errors.Join(err, c.sp.Close())
	}
	if c.ap != nil {
		err = errors.Join(err, c.ap.Close())
	}
	if c.admin != nil {
		err = errors.Join(err, c.admin.Close())
	}
	if c.cli != nil {
		err = errors.Join(err, c.cli.Close())
	}
	return err
}

func (c *Client) Ping(ctx context.Context) error {
	if c.admin == nil {
		return nil
	}
	_, err := c.admin.Controller()
	return err
}

// -----------------------------------------------------------------------------
// Producer (sync) and ConsumerGroup
// -----------------------------------------------------------------------------

type Header = sarama.RecordHeader

type Producer interface {
	Send(
		ctx context.Context,
		topic string,
		key []byte,
		value []byte,
		headers ...Header,
	) (partition int32, offset int64, err error)
}

func (c *Client) SyncProducer() (sarama.SyncProducer, error) {
	if c.sp != nil {
		return c.sp, nil
	}
	if c.cli == nil {
		return nil, fmt.Errorf("client not initialized")
	}
	sp, err := sarama.NewSyncProducerFromClient(c.cli)
	if err != nil {
		return nil, err
	}
	c.sp = sp
	return sp, nil
}

func (c *Client) Send(ctx context.Context, topic string, key, value []byte, headers ...Header) (int32, int64, error) {
	sp, err := c.SyncProducer()
	if err != nil {
		return 0, 0, err
	}
	msg := &sarama.ProducerMessage{Topic: topic, Key: sarama.ByteEncoder(key), Value: sarama.ByteEncoder(value)}
	if len(headers) > 0 {
		msg.Headers = headers
	}
	return sp.SendMessage(msg)
}

// -----------------------------------------------------------------------------
// Consumer runner: Queue (shared group) / PubSub (unique group) + semantics
// -----------------------------------------------------------------------------

type Semantics int

const (
	SemAtLeastOnce Semantics = iota
	SemAtMostOnce
	SemExactlyOnce // requires transactions (see TxnClient)
)

type ConsumeOptions struct {
	Group       string
	Concurrency int
	Semantics   Semantics
	Backoff     time.Duration
}

type HandleFunc func(ctx context.Context, m *sarama.ConsumerMessage) error

func (c *Client) RunConsumer(ctx context.Context, topics []string, opt ConsumeOptions, h HandleFunc) error {
	if opt.Concurrency <= 0 {
		opt.Concurrency = 1
	}
	if opt.Backoff <= 0 {
		opt.Backoff = 200 * time.Millisecond
	}

	cg, err := sarama.NewConsumerGroupFromClient(opt.Group, c.cli)
	if err != nil {
		return err
	}
	defer cg.Close()

	wg := new(sync.WaitGroup)
	wg.Add(opt.Concurrency)
	errCh := make(chan error, opt.Concurrency)

	handler := &consumerGroupHandler{sem: opt.Semantics, h: h}

	for i := 0; i < opt.Concurrency; i++ {
		go func() {
			defer wg.Done()
			for {
				if err := cg.Consume(ctx, topics, handler); err != nil {
					errCh <- err
					return
				}
				if ctx.Err() != nil {
					return
				}
			}
		}()
	}

	go func() { wg.Wait(); close(errCh) }()
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

type consumerGroupHandler struct {
	sem Semantics
	h   HandleFunc
}

func (h *consumerGroupHandler) Setup(s sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(s sarama.ConsumerGroupSession) error { return nil }
func (h *consumerGroupHandler) ConsumeClaim(s sarama.ConsumerGroupSession, c sarama.ConsumerGroupClaim) error {
	for m := range c.Messages() {
		switch h.sem {
		case SemAtMostOnce:
			s.MarkMessage(m, "") // commit BEFORE processing
			if err := h.h(context.Background(), m); err != nil {
				return err
			}
		default: // at-least-once base
			if err := h.h(context.Background(), m); err != nil {
				return err
			}
			s.MarkMessage(m, "") // commit AFTER processing
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// Request/Reply helper (correlation-id + reply-to header)
// -----------------------------------------------------------------------------

const (
	HdrCorrelationID = "x-correlation-id"
	HdrReplyTo       = "x-reply-to"
	HdrContentType   = "content-type"
)

func randHex(n int) string { b := make([]byte, n); _, _ = rand.Read(b); return fmt.Sprintf("%x", b) }

func (c *Client) Request(
	ctx context.Context,
	requestTopic, replyTopic string,
	key, value []byte,
	timeout time.Duration,
) (*sarama.ConsumerMessage, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	corr := randHex(12)
	hdrs := []sarama.RecordHeader{
		{Key: []byte(HdrCorrelationID), Value: []byte(corr)},
		{Key: []byte(HdrReplyTo), Value: []byte(replyTopic)},
	}
	if _, _, err := c.Send(ctx, requestTopic, key, value, hdrs...); err != nil {
		return nil, err
	}

	group := fmt.Sprintf("podzone-rr-%s", randHex(6))
	cg, err := sarama.NewConsumerGroupFromClient(group, c.cli)
	if err != nil {
		return nil, err
	}
	defer cg.Close()

	var (
		mu    sync.Mutex
		found *sarama.ConsumerMessage
		done  = make(chan struct{})
	)

	handler := &consumerGroupHandler{
		sem: SemAtLeastOnce,
		h: func(ctx context.Context, m *sarama.ConsumerMessage) error {
			for _, h := range m.Headers {
				if strings.EqualFold(string(h.Key), HdrCorrelationID) && string(h.Value) == corr {
					mu.Lock()
					found = m
					mu.Unlock()
					close(done)
					break
				}
			}
			return nil
		},
	}

	go func() { _ = cg.Consume(ctx, []string{replyTopic}, handler) }()
	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		return found, nil
	case <-time.After(timeout):
		return nil, context.DeadlineExceeded
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// -----------------------------------------------------------------------------
// Transactions (Exactly-once) — skeleton
// -----------------------------------------------------------------------------

// To enable transactions:
//   sc.Producer.Idempotent = true
//   sc.Producer.RequiredAcks = sarama.WaitForAll
//   sc.Net.MaxOpenRequests = 1
//   sc.Producer.Transaction.ID = "podzone-<stable-id>"
// Downstream consumers that must ignore aborted records:
//   sc.Consumer.IsolationLevel = sarama.ReadCommitted

// TxnClient wraps a transactional AsyncProducer. You must drain Successes/Errors.

type TxnClient struct{ prod sarama.AsyncProducer }

func (c *Client) NewTxnClient() (*TxnClient, error) {
	if c.scfg.Producer.Transaction.ID == "" || !c.scfg.Producer.Idempotent {
		return nil, fmt.Errorf("transactions require Idempotent=true and Transaction.ID set")
	}
	ap, err := sarama.NewAsyncProducerFromClient(c.cli)
	if err != nil {
		return nil, err
	}
	return &TxnClient{prod: ap}, nil
}

func (t *TxnClient) Close() error                    { return t.prod.Close() }
func (t *TxnClient) Begin(ctx context.Context) error { return t.prod.BeginTxn() }
func (t *TxnClient) Abort(ctx context.Context) error { return t.prod.AbortTxn() }

func (t *TxnClient) Produce(ctx context.Context, topic string, key, value []byte, headers ...Header) error {
	t.prod.Input() <- &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.ByteEncoder(key),
		Value:   sarama.ByteEncoder(value),
		Headers: headers,
	}
	return nil
}

// SendOffsetsAndCommit should:
//   - call the appropriate SendOffsetsToTxn/Offset API with the current consumer group metadata
//   - then CommitTxn(). If any step fails, AbortTxn().
//
// Exact call signature differs by Sarama revision; wire your ConsumerGroupSession fields here.
func (t *TxnClient) SendOffsetsAndCommit(
	ctx context.Context, /* session sarama.ConsumerGroupSession, meta ... */
) error {
	// TODO: t.prod.SendOffsetsToTxn(ctx, groupMetadata, offsets)
	return t.prod.CommitTxn()
}
