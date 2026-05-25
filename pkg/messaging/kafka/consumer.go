package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"

	"github.com/tuannm99/podzone/pkg/messaging"
	"github.com/tuannm99/podzone/pkg/pdkafka"
)

type ConsumerOptions struct {
	Publisher        messaging.Publisher
	RetryPolicy      messaging.RetryPolicy
	DeadLetterPolicy messaging.DeadLetterPolicy
	TopicStrategy    messaging.TopicStrategy
	Classifier       messaging.ErrorClassifier
	Middlewares      []messaging.Middleware
	ConsumerName     string
	Observer         messaging.Observer
	Now              func() time.Time
}

type Consumer struct {
	runner  pdkafka.ConsumerGroupRunner
	topics  []string
	handler messaging.Handler
	opts    ConsumerOptions
}

var _ messaging.Consumer = (*Consumer)(nil)

func NewConsumer(runner pdkafka.ConsumerGroupRunner, topics []string, handler messaging.Handler) *Consumer {
	return NewConsumerWithOptions(runner, topics, handler, ConsumerOptions{})
}

func NewConsumerWithOptions(
	runner pdkafka.ConsumerGroupRunner,
	topics []string,
	handler messaging.Handler,
	opts ConsumerOptions,
) *Consumer {
	if opts.Classifier == nil {
		opts.Classifier = messaging.DefaultErrorClassifier()
	}
	if opts.Now == nil {
		opts.Now = func() time.Time { return time.Now().UTC() }
	}
	if opts.TopicStrategy == (messaging.TopicStrategy{}) {
		opts.TopicStrategy = messaging.DefaultTopicStrategy()
	}
	if handler != nil && len(opts.Middlewares) > 0 {
		handler = messaging.Chain(handler, opts.Middlewares...)
	}
	return &Consumer{
		runner:  runner,
		topics:  append([]string(nil), topics...),
		handler: handler,
		opts:    opts,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	return c.runner.Run(ctx, c.topics, &consumerGroupHandler{
		handler: c.handler,
		opts:    c.opts,
	})
}

type consumerGroupHandler struct {
	handler messaging.Handler
	opts    ConsumerOptions
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for {
		select {
		case <-session.Context().Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			if err := h.consumeMessage(session, msg); err != nil {
				return err
			}
		}
	}
}

func (h *consumerGroupHandler) consumeMessage(
	session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage,
) error {
	env, err := h.decodeEnvelope(msg)
	if err != nil {
		return err
	}
	if err := h.waitUntilReady(session.Context(), env); err != nil {
		return err
	}
	if h.handler == nil {
		h.observe(session.Context(), msg, env, messaging.FailureActionDrop, "no_handler", nil)
		session.MarkMessage(msg, "")
		return nil
	}
	if err := h.handler.Handle(session.Context(), env); err != nil {
		return h.handleFailure(session, msg, env, err)
	}
	h.observe(session.Context(), msg, env, messaging.FailureActionReturn, "handled", nil)
	session.MarkMessage(msg, "")
	return nil
}

func (h *consumerGroupHandler) waitUntilReady(ctx context.Context, env messaging.Envelope) error {
	metadata := messaging.ReadDeliveryMetadata(env)
	if metadata.NextAttemptAt.IsZero() {
		return nil
	}
	delay := time.Until(metadata.NextAttemptAt)
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (h *consumerGroupHandler) decodeEnvelope(msg *sarama.ConsumerMessage) (messaging.Envelope, error) {
	var env messaging.Envelope
	if err := json.Unmarshal(msg.Value, &env); err != nil {
		return messaging.Envelope{}, fmt.Errorf("unmarshal envelope: %w", err)
	}
	recordHeaders := pdkafka.FromRecordHeaders(msg.Headers)
	if len(recordHeaders) > 0 {
		clone := env.Clone()
		if clone.Headers == nil {
			clone.Headers = make(map[string]string, len(recordHeaders))
		}
		for k, v := range recordHeaders {
			if _, exists := clone.Headers[k]; !exists {
				clone.Headers[k] = v
			}
		}
		env = clone
	}
	return env, nil
}

func (h *consumerGroupHandler) handleFailure(
	session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage,
	env messaging.Envelope,
	err error,
) error {
	classifier := h.opts.Classifier
	if classifier == nil {
		classifier = messaging.DefaultErrorClassifier()
	}
	classification := classifier.Classify(session.Context(), env, err)
	switch classification.Action {
	case messaging.FailureActionDrop:
		h.observe(session.Context(), msg, env, classification.Action, classification.Reason, classification.Err)
		session.MarkMessage(msg, "")
		return nil
	case messaging.FailureActionRetry:
		return h.publishRetry(session, msg, env, classification)
	case messaging.FailureActionDeadLetter:
		return h.publishDeadLetter(session, msg, env, classification)
	case messaging.FailureActionReturn, "":
		return err
	default:
		return err
	}
}

func (h *consumerGroupHandler) publishRetry(
	session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage,
	env messaging.Envelope,
	classification messaging.FailureClassification,
) error {
	if h.opts.Publisher == nil {
		return classification.Err
	}

	metadata := messaging.ReadDeliveryMetadata(env)
	nextAttempt := metadata.Attempt + 1
	policy := h.opts.RetryPolicy.Normalize()
	now := h.opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	if policy.Exhausted(nextAttempt) {
		return h.publishDeadLetter(session, msg, env, messaging.FailureClassification{
			Action: messaging.FailureActionDeadLetter,
			Reason: classification.Reason,
			Err:    classification.Err,
		})
	}

	original := originalTopic(metadata.OriginalTopic, msg.Topic)
	metadata.Attempt = nextAttempt
	metadata.MaxAttempts = policy.MaxAttempts
	metadata.OriginalTopic = original
	metadata.LastError = errorText(classification.Err)
	metadata.ConsumerName = h.opts.ConsumerName
	metadata.NextAttemptAt = now().Add(policy.NextDelay(nextAttempt))
	retryEnv := messaging.WithDeliveryMetadata(env, metadata)
	retryTopic := h.opts.TopicStrategy.RetryTopic(original, nextAttempt)
	if err := h.opts.Publisher.Publish(session.Context(), retryTopic, msgKey(msg), retryEnv); err != nil {
		return fmt.Errorf("publish retry message: %w", err)
	}
	h.observe(session.Context(), msg, retryEnv, messaging.FailureActionRetry, classification.Reason, classification.Err)
	session.MarkMessage(msg, "")
	return nil
}

func (h *consumerGroupHandler) publishDeadLetter(
	session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage,
	env messaging.Envelope,
	classification messaging.FailureClassification,
) error {
	if h.opts.Publisher == nil {
		return classification.Err
	}
	metadata := messaging.ReadDeliveryMetadata(env)
	original := originalTopic(metadata.OriginalTopic, msg.Topic)
	metadata.OriginalTopic = original
	metadata.LastError = errorText(classification.Err)
	metadata.DeadLetterReason = classification.Reason
	metadata.ConsumerName = h.opts.ConsumerName
	metadata.NextAttemptAt = time.Time{}
	deadEnv := messaging.WithDeliveryMetadata(env, metadata)
	deadTopic := h.opts.DeadLetterPolicy.TopicFor(original)
	if deadTopic == "" {
		deadTopic = h.opts.TopicStrategy.DeadLetterTopic(original)
	}
	if err := h.opts.Publisher.Publish(session.Context(), deadTopic, msgKey(msg), deadEnv); err != nil {
		return fmt.Errorf("publish dead letter message: %w", err)
	}
	h.observe(
		session.Context(),
		msg,
		deadEnv,
		messaging.FailureActionDeadLetter,
		classification.Reason,
		classification.Err,
	)
	session.MarkMessage(msg, "")
	return nil
}

func (h *consumerGroupHandler) observe(
	ctx context.Context,
	msg *sarama.ConsumerMessage,
	env messaging.Envelope,
	action messaging.FailureAction,
	reason string,
	err error,
) {
	if h.opts.Observer == nil {
		return
	}
	h.opts.Observer.Observe(ctx, messaging.DeliveryEvent{
		ConsumerName: h.opts.ConsumerName,
		Topic:        msg.Topic,
		Key:          msgKey(msg),
		Envelope:     env,
		Action:       action,
		Reason:       reason,
		Err:          err,
	})
}

func originalTopic(existing string, fallback string) string {
	if existing != "" {
		return existing
	}
	return fallback
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func msgKey(msg *sarama.ConsumerMessage) string {
	if msg == nil || msg.Key == nil {
		return ""
	}
	return string(msg.Key)
}
