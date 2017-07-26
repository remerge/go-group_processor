package groupprocessor

import (
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
)

type testLoadSaver struct {
	DefaultLoadSaver
	channel chan string
}

func (ls *testLoadSaver) Save(p Processable) error {
	tp := p.(*DefaultProcessable)
	ls.channel <- string(tp.Msg().Value)
	return nil
}

func assertEqual(t *testing.T, a interface{}, b interface{}, message string, args ...interface{}) {
	if a != b {
		t.Fatalf(message, args)
	}
}

func TestGroupProcessor(t *testing.T) {
	tls := &testLoadSaver{
		channel: make(chan string),
	}

	gp := &GroupProcessor{
		Name:          "gp",
		Brokers:       "localhost:9092",
		Topic:         "test",
		NumLoadWorker: 4,
		NumSaveWorker: 4,
		LoadSaver:     tls,
	}

	if err := gp.New(); err != nil {
		t.Errorf("Unexpected error in gp.New: %v", err)
		return
	}

	gp.Run()

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, nil)
	assertEqual(t, err, nil, "Unexpected error in NewSyncProducer: %v", err)

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: "test",
		Value: sarama.StringEncoder("test"),
	})

	assertEqual(t, err, nil, "Unexpected error in SendMessage: %v", err)

	var msg string

L:
	for {
		select {
		// drain channel
		case msg = <-tls.channel:
			fmt.Printf("msg=%#v\n", msg)
		case <-time.After(100 * time.Millisecond):
			gp.Close()
			break L
		}
	}

	assertEqual(t, msg, "test", "expected message to equal \"true\", got %#v", msg)
}

func TestGroupProcessor_with_CustomConfig(t *testing.T) {
	tls := &testLoadSaver{
		channel: make(chan string),
	}

	config := sarama.NewConfig()
	config.ClientID = "TEST"                             // ClientID will always be overridden
	config.Version = sarama.V0_8_2_0                     // Version will always be overridden
	config.Consumer.MaxProcessingTime = 30 * time.Second // everything will be set
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	gp := &GroupProcessor{
		Name:              "gp",
		Brokers:           "localhost:9092",
		Topic:             "test",
		NumLoadWorker:     4,
		NumSaveWorker:     4,
		LoadSaver:         tls,
		CustomKafkaConfig: config,
	}

	if err := gp.New(); err != nil {
		t.Errorf("Unexpected error in gp.New: %v", err)
		return
	}

	assertEqual(t, gp.kafka.config.ClientID, "gp.test", "ClientID should always be created as <Name>.<Topic>")
	assertEqual(t, gp.kafka.config.Version, sarama.V0_10_0_0, "Version will always be set to V0_10_0_0")
	assertEqual(t, gp.kafka.config.Consumer.MaxProcessingTime, 30*time.Second, "MaxProcessingTime should be 30s")
	assertEqual(t, gp.kafka.config.Consumer.Offsets.Initial, sarama.OffsetNewest, "Offsets.Initial should be OffsetNewest")

	gp.Run()

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, nil)
	assertEqual(t, err, nil, "Unexpected error in NewSyncProducer: %v", err)

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: "test",
		Value: sarama.StringEncoder("test"),
	})

	assertEqual(t, err, nil, "Unexpected error in SendMessage: %v", err)

	var msg string

L:
	for {
		select {
		// drain channel
		case msg = <-tls.channel:
			fmt.Printf("msg=%#v\n", msg)
		case <-time.After(100 * time.Millisecond):
			gp.Close()
			break L
		}
	}

	assertEqual(t, msg, "test", "expected message to equal \"true\", got %#v", msg)

}
