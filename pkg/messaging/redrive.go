package messaging

import "context"

type RedriveRequest struct {
	SourceTopic string
	TargetTopic string
	Key         string
	Envelope    Envelope
}

type Redriver struct {
	publisher Publisher
}

func NewRedriver(publisher Publisher) *Redriver {
	return &Redriver{publisher: publisher}
}

func (r *Redriver) Redrive(ctx context.Context, request RedriveRequest) error {
	env := request.Envelope.Clone()
	metadata := ReadDeliveryMetadata(env)
	metadata.OriginalTopic = request.SourceTopic
	metadata.RedriveCount++
	env = WithDeliveryMetadata(env, metadata)
	return r.publisher.Publish(ctx, request.TargetTopic, request.Key, env)
}

func (r *Redriver) RedriveBatch(ctx context.Context, requests []RedriveRequest) error {
	items := make([]PublishRequest, 0, len(requests))
	for _, request := range requests {
		env := request.Envelope.Clone()
		metadata := ReadDeliveryMetadata(env)
		metadata.OriginalTopic = request.SourceTopic
		metadata.RedriveCount++
		env = WithDeliveryMetadata(env, metadata)
		items = append(items, PublishRequest{
			Topic: request.TargetTopic,
			Key:   request.Key,
			Msg:   env,
		})
	}
	return r.publisher.PublishBatch(ctx, "", items)
}
