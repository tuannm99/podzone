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
