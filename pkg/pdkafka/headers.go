package pdkafka

import "github.com/IBM/sarama"

func ToRecordHeaders(headers map[string]string) []sarama.RecordHeader {
	if len(headers) == 0 {
		return nil
	}
	out := make([]sarama.RecordHeader, 0, len(headers))
	for k, v := range headers {
		out = append(out, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}
	return out
}

func FromRecordHeaders(headers []*sarama.RecordHeader) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(headers))
	for _, header := range headers {
		if header == nil {
			continue
		}
		out[string(header.Key)] = string(header.Value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
