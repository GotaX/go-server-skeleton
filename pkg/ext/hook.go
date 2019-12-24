package ext

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

const (
	batchSize     = 100
	flushInterval = 3 * time.Second
)

var (
	keyMessage = "message"
	keyLevel   = "level"
)

type Message struct {
	Time     time.Time
	Contents map[string]string
}

type LogStoreHook struct {
	mu        *sync.Mutex
	store     *sls.LogStore
	topic     string
	source    string
	extra     map[string]interface{}
	messages  []Message
	flushTime time.Time
}

func NewLogStoreHook(endpoint, key, secret, projectName, storeName, topic, source string, extra map[string]interface{}) (*LogStoreHook, error) {
	client := &sls.Client{
		Endpoint:        endpoint,
		AccessKeyID:     key,
		AccessKeySecret: secret,
	}

	project, err := client.GetProject(projectName)
	if err != nil {
		return nil, xerrors.Errorf("unknown LogStore project %v: %v", projectName, err)
	}

	store, err := project.GetLogStore(storeName)
	if err != nil {
		return nil, xerrors.Errorf("unknown LogStore name %v: %v", storeName, err)
	}

	return &LogStoreHook{
		mu:        &sync.Mutex{},
		store:     store,
		topic:     topic,
		source:    source,
		extra:     extra,
		messages:  make([]Message, 0, batchSize),
		flushTime: time.Now(),
	}, nil
}

func (h *LogStoreHook) Fire(e *logrus.Entry) error {
	m := Message{
		Time:     e.Time,
		Contents: make(map[string]string, len(e.Data)+len(h.extra)+2),
	}

	m.Contents[keyMessage] = e.Message
	m.Contents[keyLevel] = strconv.Itoa(logrusToSyslog(e.Level))

	for k, v := range h.extra {
		m.Contents[k] = toString(v)
	}
	for k, v := range e.Data {
		m.Contents[k] = toString(v)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.messages = append(h.messages, m)

	if err := h.Flush(false); err != nil {
		return xerrors.Errorf("while flush log messages: %v", err)
	}
	return nil
}

func (h *LogStoreHook) Flush(force bool) error {
	if len(h.messages) == 0 || !force && len(h.messages) < batchSize && time.Since(h.flushTime) < flushInterval {
		return nil
	}

	size := len(h.messages)
	messages := make([]*sls.Log, size)
	for i, m := range h.messages {
		contents := make([]*sls.LogContent, 0, len(m.Contents))
		for k, v := range m.Contents {
			contents = append(contents, &sls.LogContent{
				Key:   proto.String(k),
				Value: proto.String(v),
			})
		}
		messages[i] = &sls.Log{
			Time:     proto.Uint32(uint32(m.Time.Unix())),
			Contents: contents,
		}
	}

	lg := &sls.LogGroup{
		Topic:  &h.topic,
		Source: &h.source,
		Logs:   messages,
	}

	if err := h.store.PutLogs(lg); err != nil {
		for _, message := range lg.Logs {
			if err := h.store.PutLogs(&sls.LogGroup{
				Topic:  &h.topic,
				Source: &h.source,
				Logs:   []*sls.Log{message},
			}); err != nil {
				log.Println("Discard log: ", err)
				continue
			}
		}
	}

	log.Printf("Flush %v messages to LogStore (%v)", size, h.store.Name)

	h.flushTime = time.Now()
	h.messages = h.messages[:0]
	return nil
}

func (h *LogStoreHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel}
}

func toString(v interface{}) string {
	switch vv := v.(type) {
	case string:
		return vv
	default:
		return fmt.Sprintf("%v", v)
	}
}

func logrusToSyslog(level logrus.Level) int {
	switch level {
	case logrus.PanicLevel:
		return 0
	case logrus.FatalLevel:
		return 2
	case logrus.ErrorLevel:
		return 3
	case logrus.WarnLevel:
		return 4
	case logrus.InfoLevel:
		return 6
	case logrus.DebugLevel:
		return 7
	case logrus.TraceLevel:
		return 8
	default:
		return -1
	}
}
