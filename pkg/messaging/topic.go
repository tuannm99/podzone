package messaging

import "fmt"

func Topic(service string, stream string) string {
	return fmt.Sprintf("podzone.%s.%s", service, stream)
}
