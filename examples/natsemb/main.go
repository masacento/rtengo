package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"

	rten "github.com/masacento/rtengo"
)

const (
	serviceName    = "rtengo-embedding"
	serviceVersion = "1.0.0"
)

type embedRequest struct {
	Text string `json:"text"`
}

type embedResponse struct {
	Embedding []float32 `json:"embedding"`
	Dimension int       `json:"dimension"`
}

// embeddingHandler serializes inference because a model and its WASM runtime
// are shared by all requests handled by this service instance.
type embeddingHandler struct {
	embedder  *rten.Embedder
	tokenizer rten.Tokenizer
	debug     bool
	mu        sync.Mutex
}

func (h *embeddingHandler) handle(req micro.Request) {
	requestStarted := time.Now()

	var input embedRequest
	if err := json.Unmarshal(req.Data(), &input); err != nil {
		respondError(req, "400", "request must be JSON with a text field", err)
		return
	}
	if input.Text == "" {
		respondError(req, "400", "text must not be empty", nil)
		return
	}

	waitStarted := time.Now()
	h.mu.Lock()
	waitDuration := time.Since(waitStarted)

	tokenCount := -1
	tokenizeStarted := time.Now()
	if h.debug {
		encoding, tokenizeErr := h.tokenizer.Encode(input.Text, true)
		if tokenizeErr != nil {
			log.Printf("debug: count input tokens: %v", tokenizeErr)
		} else {
			tokenCount = len(encoding.IDs)
		}
	}
	tokenizeDuration := time.Since(tokenizeStarted)

	inferenceStarted := time.Now()
	embedding, err := h.embedder.Embed(context.Background(), input.Text)
	inferenceDuration := time.Since(inferenceStarted)
	h.mu.Unlock()
	if err != nil {
		respondError(req, "500", "embedding failed", err)
		return
	}

	payload, err := json.Marshal(embedResponse{
		Embedding: embedding,
		Dimension: len(embedding),
	})
	if err != nil {
		respondError(req, "500", "response encoding failed", err)
		return
	}
	if err := req.Respond(payload); err != nil {
		log.Printf("failed to send response: %v", err)
	}
	if h.debug {
		log.Printf(
			"debug: embedding tokens=%d text_bytes=%d wait=%s tokenize=%s inference=%s total=%s dimension=%d",
			tokenCount,
			len(input.Text),
			waitDuration.Truncate(time.Microsecond),
			tokenizeDuration.Truncate(time.Microsecond),
			inferenceDuration.Truncate(time.Microsecond),
			time.Since(requestStarted).Truncate(time.Microsecond),
			len(embedding),
		)
	}
}

func respondError(req micro.Request, code, message string, cause error) {
	if cause != nil {
		log.Printf("%s: %v", message, cause)
	}
	if err := req.Error(code, message, nil); err != nil {
		log.Printf("failed to send error response: %v", err)
	}
}

func main() {
	modelPath := flag.String("model", "./model.rten", "path to model file")
	tokenizerPath := flag.String("tokenizer", "./tokenizer.json", "path to tokenizer file")
	natsURL := flag.String("nats", nats.DefaultURL, "NATS server URL")
	subject := flag.String("subject", "rtengo.embed", "NATS request subject")
	debug := flag.Bool("debug", false, "log token counts and request timings")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("loading model %q", *modelPath)
	started := time.Now()
	runtime, err := rten.NewRuntime(ctx)
	if err != nil {
		log.Fatalf("create runtime: %v", err)
	}
	defer runtime.Close(context.Background())

	modelBytes, err := os.ReadFile(*modelPath)
	if err != nil {
		log.Fatalf("read model: %v", err)
	}
	model, err := runtime.LoadModel(ctx, modelBytes)
	if err != nil {
		log.Fatalf("load model: %v", err)
	}
	defer model.Close(context.Background())

	tokenizer, err := rten.NewTokenizerFromFile(*tokenizerPath)
	if err != nil {
		log.Fatalf("load tokenizer: %v", err)
	}
	embedder, err := rten.NewEmbedder(ctx, runtime, model, tokenizer)
	if err != nil {
		log.Fatalf("create embedder: %v", err)
	}
	log.Printf("model ready in %s", time.Since(started).Truncate(time.Millisecond))

	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Fatalf("connect to NATS: %v", err)
	}
	defer nc.Close()

	handler := &embeddingHandler{
		embedder:  embedder,
		tokenizer: tokenizer,
		debug:     *debug,
	}
	service, err := micro.AddService(nc, micro.Config{
		Name:        serviceName,
		Version:     serviceVersion,
		Description: "Generate text embeddings with RTen",
		Endpoint: &micro.EndpointConfig{
			Subject: *subject,
			Handler: micro.HandlerFunc(handler.handle),
		},
	})
	if err != nil {
		log.Fatalf("start NATS service: %v", err)
	}

	log.Printf("service %s %s listening on %q via %s", serviceName, serviceVersion, *subject, *natsURL)
	<-ctx.Done()
	log.Print("shutting down")
	if err := service.Stop(); err != nil {
		log.Printf("stop service: %v", err)
	}
	if err := nc.Drain(); err != nil {
		log.Printf("drain NATS connection: %v", err)
	}
}
