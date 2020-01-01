package ext

import (
	"log"
	"strconv"
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type AsyncLogStoreHook struct {
	remote       *sls.LogStore
	chMsg        chan Message
	chQuit       chan struct{}
	bufferSize   int
	interval     time.Duration
	flushTimeout time.Duration
	topic        string
	source       string
	extra        map[string]interface{}
}

func NewAsyncLogStoreHook(c LogStoreConfig) (logrus.Hook, error) {
	client := &sls.Client{
		Endpoint:        c.Endpoint,
		AccessKeyID:     c.AccessKey,
		AccessKeySecret: c.AccessSecret,
	}

	project, err := client.GetProject(c.Project)
	if err != nil {
		return nil, xerrors.Errorf("unknown LogStore project %v: %v", c.Project, err)
	}

	store, err := project.GetLogStore(c.Store)
	if err != nil {
		return nil, xerrors.Errorf("unknown LogStore name %v: %v", c.Store, err)
	}
	obj := &AsyncLogStoreHook{
		remote:       store,
		chMsg:        make(chan Message, batchSize),
		chQuit:       make(chan struct{}),
		bufferSize:   batchSize,
		interval:     flushInterval,
		flushTimeout: flushTimeout,
		topic:        c.Topic,
		source:       c.Source,
		extra:        c.Extra,
	}
	go obj.start()
	return obj, nil
}

func (h *AsyncLogStoreHook) start() {
	messages := make([]Message, 0, h.bufferSize)
	lastFlushTime := time.Now()

	needFlush := func() bool {
		return len(messages) >= h.bufferSize || time.Since(lastFlushTime) >= h.interval
	}

	flush := func() {
		if len(messages) == 0 {
			return
		}

		startTime := time.Now()

		h.flush(messages)

		log.Printf("[%v] Flush %d messages to %q",
			time.Since(startTime).Truncate(time.Millisecond),
			len(messages), h.remote.Name)

		messages = messages[:0]
		lastFlushTime = time.Now()
	}

Loop:
	for {
		select {
		case <-time.After(h.interval / 10):
			// Check flush
		case message, ok := <-h.chMsg:
			if !ok {
				break Loop
			}
			messages = append(messages, message)
		}

		if needFlush() {
			flush()
		}
	}

	flush()
	close(chQuit)
}

func (h *AsyncLogStoreHook) Close() error {
	close(h.chMsg)
	<-chQuit
	return nil
}

func (h *AsyncLogStoreHook) Fire(entry *logrus.Entry) error {
	select {
	case h.chMsg <- h.toMessage(entry):
	case <-time.After(h.flushTimeout):
	}
	return nil
}

func (h *AsyncLogStoreHook) toMessage(entry *logrus.Entry) Message {
	m := Message{
		Time:     entry.Time,
		Contents: make(map[string]string, len(entry.Data)+len(h.extra)+2),
	}

	m.Contents[keyMessage] = entry.Message
	m.Contents[keyLevel] = strconv.Itoa(logrusToSyslog(entry.Level))

	for k, v := range h.extra {
		m.Contents[k] = toString(v)
	}
	for k, v := range entry.Data {
		m.Contents[k] = toString(v)
	}
	return m
}

func (h *AsyncLogStoreHook) flush(messages []Message) {
	msgs := make([]*sls.Log, len(messages))
	for i, m := range messages {
		contents := make([]*sls.LogContent, 0, len(m.Contents))
		for k, v := range m.Contents {
			contents = append(contents, &sls.LogContent{
				Key:   proto.String(k),
				Value: proto.String(v),
			})
		}
		msgs[i] = &sls.Log{
			Time:     proto.Uint32(uint32(m.Time.Unix())),
			Contents: contents,
		}
	}

	lg := &sls.LogGroup{
		Topic:  &h.topic,
		Source: &h.source,
		Logs:   msgs,
	}

	if err := h.remote.PutLogs(lg); err != nil {
		for _, message := range lg.Logs {
			if err := h.remote.PutLogs(&sls.LogGroup{
				Topic:  &h.topic,
				Source: &h.source,
				Logs:   []*sls.Log{message},
			}); err != nil {
				log.Println("Discard log: ", err)
				continue
			}
		}
	}
}

func (h *AsyncLogStoreHook) Levels() []logrus.Level { return HookedLevels }
