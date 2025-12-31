package saramaProducerx

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gitee.com/hgg_test/pkg_tool/v2/channelx/messageQueuex"
	"github.com/IBM/sarama"
)

// 定义生产者类型枚举
type ProducerType int

const (
	ProducerTypeSync  ProducerType = 0
	ProducerTypeAsync ProducerType = 1
)

// 定义通用生产者接口
type Producer interface {
	// SendMessage 发送消息
	SendMessage(ctx context.Context, keyOrTopic messageQueuex.Tp, value []byte) error
	// Close 关闭生产者，释放资源
	Close() error
	// Type 返回生产者类型
	Type() ProducerType
}

// AsyncResultHandler 异步结果处理器接口
type AsyncResultHandler interface {
	// HandleSuccess 处理成功发送的消息
	HandleSuccess(success *sarama.ProducerMessage)
	// HandleError 处理发送失败的消息
	HandleError(err *sarama.ProducerError)
}

// DefaultAsyncResultHandler 默认的异步结果处理器
type DefaultAsyncResultHandler struct{}

func (h *DefaultAsyncResultHandler) HandleSuccess(success *sarama.ProducerMessage) {
	// 默认实现：记录日志
	// 实际使用中可以自定义实现，例如重试、监控等
}

func (h *DefaultAsyncResultHandler) HandleError(err *sarama.ProducerError) {
	// 默认实现：记录错误日志
	// 实际使用中可以自定义实现，例如重试、告警等
}

// SyncProducer 同步生产者实现
type SyncProducer struct {
	producer sarama.SyncProducer
	config   *sarama.Config
}

// NewSyncProducer 创建同步生产者
func NewSyncProducer(brokers []string, config *sarama.Config) (*SyncProducer, error) {
	if config == nil {
		config = sarama.NewConfig()
	}
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建同步生产者失败: %w", err)
	}

	return &SyncProducer{
		producer: producer,
		config:   config,
	}, nil
}

// SendMessage 发送消息（同步）
func (p *SyncProducer) SendMessage(ctx context.Context, keyOrTopic messageQueuex.Tp, value []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	msg := &sarama.ProducerMessage{
		Topic: keyOrTopic.Topic,
		Key:   sarama.StringEncoder(keyOrTopic.Key),
		Value: sarama.StringEncoder(value),
	}

	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	return nil
}

// Close 关闭同步生产者
func (p *SyncProducer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// Type 返回生产者类型
func (p *SyncProducer) Type() ProducerType {
	return ProducerTypeSync
}

// AsyncProducer 异步生产者实现
type AsyncProducer struct {
	producer    sarama.AsyncProducer
	config      *sarama.Config
	resultChan  chan interface{}
	stopChan    chan struct{}
	stoppedChan chan struct{}
	wg          sync.WaitGroup
	handler     AsyncResultHandler
}

// NewAsyncProducer 创建异步生产者
func NewAsyncProducer(brokers []string, config *sarama.Config, handler AsyncResultHandler) (*AsyncProducer, error) {
	if config == nil {
		config = sarama.NewConfig()
	}
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建异步生产者失败: %w", err)
	}

	if handler == nil {
		handler = &DefaultAsyncResultHandler{}
	}

	asyncProducer := &AsyncProducer{
		producer:    producer,
		config:      config,
		resultChan:  make(chan interface{}, 100),
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
		handler:     handler,
	}

	// 启动结果处理goroutine
	asyncProducer.wg.Add(1)
	go asyncProducer.handleResults()

	return asyncProducer, nil
}

// SendMessage 发送消息（异步）
func (p *AsyncProducer) SendMessage(ctx context.Context, keyOrTopic messageQueuex.Tp, value []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	msg := &sarama.ProducerMessage{
		Topic: keyOrTopic.Topic,
		Key:   sarama.StringEncoder(keyOrTopic.Key),
		Value: sarama.StringEncoder(value),
	}

	// 异步发送消息
	select {
	case p.producer.Input() <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.stopChan:
		return errors.New("生产者已关闭")
	}
}

// handleResults 处理异步发送结果
func (p *AsyncProducer) handleResults() {
	defer p.wg.Done()
	defer close(p.stoppedChan)

	for {
		select {
		case success, ok := <-p.producer.Successes():
			if !ok {
				// 通道已关闭，退出
				return
			}
			p.handler.HandleSuccess(success)

		case err, ok := <-p.producer.Errors():
			if !ok {
				// 通道已关闭，退出
				return
			}
			p.handler.HandleError(err)

		case <-p.stopChan:
			// 收到停止信号，关闭生产者并退出
			if err := p.producer.Close(); err != nil {
				// 记录关闭错误但不返回，继续执行退出流程
			}
			return
		}
	}
}

// Close 关闭异步生产者，等待所有goroutine退出
func (p *AsyncProducer) Close() error {
	close(p.stopChan)

	// 等待结果处理goroutine退出
	p.wg.Wait()

	// 确保所有通道都已关闭
	select {
	case <-p.stoppedChan:
		// 已经停止
	default:
		close(p.stoppedChan)
	}

	return nil
}

// Type 返回生产者类型
func (p *AsyncProducer) Type() ProducerType {
	return ProducerTypeAsync
}

// BatchAsyncProducer 批量异步生产者（支持批量发送配置）
type BatchAsyncProducer struct {
	*AsyncProducer
	batchSize     int
	batchInterval time.Duration
	batchChan     chan *sarama.ProducerMessage
	batchWG       sync.WaitGroup
}

// NewBatchAsyncProducer 创建批量异步生产者
func NewBatchAsyncProducer(brokers []string, config *sarama.Config, handler AsyncResultHandler,
	batchSize int, batchInterval time.Duration) (*BatchAsyncProducer, error) {

	if config == nil {
		config = sarama.NewConfig()
	}
	config.Producer.Flush.Messages = batchSize
	config.Producer.Flush.Frequency = batchInterval

	asyncProducer, err := NewAsyncProducer(brokers, config, handler)
	if err != nil {
		return nil, err
	}

	batchProducer := &BatchAsyncProducer{
		AsyncProducer: asyncProducer,
		batchSize:     batchSize,
		batchInterval: batchInterval,
		batchChan:     make(chan *sarama.ProducerMessage, batchSize*2),
	}

	// 启动批量处理goroutine
	batchProducer.batchWG.Add(1)
	go batchProducer.batchProcessor()

	return batchProducer, nil
}

// batchProcessor 批量消息处理器
func (p *BatchAsyncProducer) batchProcessor() {
	defer p.batchWG.Done()

	ticker := time.NewTicker(p.batchInterval)
	defer ticker.Stop()

	batch := make([]*sarama.ProducerMessage, 0, p.batchSize)

	for {
		select {
		case msg := <-p.batchChan:
			if msg == nil {
				// 通道已关闭，处理剩余消息
				if len(batch) > 0 {
					p.sendBatch(batch)
				}
				return
			}

			batch = append(batch, msg)
			if len(batch) >= p.batchSize {
				p.sendBatch(batch)
				batch = batch[:0] // 重置batch
			}

		case <-ticker.C:
			if len(batch) > 0 {
				p.sendBatch(batch)
				batch = batch[:0] // 重置batch
			}

		case <-p.stopChan:
			// 处理剩余消息
			if len(batch) > 0 {
				p.sendBatch(batch)
			}
			return
		}
	}
}

// sendBatch 发送批量消息
func (p *BatchAsyncProducer) sendBatch(batch []*sarama.ProducerMessage) {
	for _, msg := range batch {
		select {
		case p.producer.Input() <- msg:
			// 成功入队
		case <-p.stopChan:
			// 生产者已关闭，停止发送
			return
		}
	}
}

// SendMessage 发送消息（批量异步）
func (p *BatchAsyncProducer) SendMessage(ctx context.Context, keyOrTopic messageQueuex.Tp, value []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	msg := &sarama.ProducerMessage{
		Topic: keyOrTopic.Topic,
		Key:   sarama.StringEncoder(keyOrTopic.Key),
		Value: sarama.StringEncoder(value),
	}

	// 将消息加入批量队列
	select {
	case p.batchChan <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-p.stopChan:
		return errors.New("生产者已关闭")
	}
}

// Close 关闭批量异步生产者
func (p *BatchAsyncProducer) Close() error {
	// 先关闭批量处理通道
	close(p.batchChan)

	// 等待批量处理器退出
	p.batchWG.Wait()

	// 调用父类的Close方法
	return p.AsyncProducer.Close()
}

// 兼容旧版本的包装器（为了向后兼容）
type SaramaProducerWrapper struct {
	producer Producer
	typ      ProducerType
}

// NewSaramaProducerWrapper 创建生产者包装器（向后兼容）
func NewSaramaProducerWrapper(producer Producer) *SaramaProducerWrapper {
	return &SaramaProducerWrapper{
		producer: producer,
		typ:      producer.Type(),
	}
}

// SendMessage 发送消息
func (w *SaramaProducerWrapper) SendMessage(ctx context.Context, keyOrTopic messageQueuex.Tp, value []byte) error {
	return w.producer.SendMessage(ctx, keyOrTopic, value)
}

// CloseProducer 关闭生产者
func (w *SaramaProducerWrapper) CloseProducer() error {
	return w.producer.Close()
}

// ProducerTyp 获取生产者类型（兼容旧版本）
func (w *SaramaProducerWrapper) ProducerTyp() uint {
	return uint(w.typ)
}

// IsAsync 检查是否为异步生产者
func (w *SaramaProducerWrapper) IsAsync() bool {
	return w.typ == ProducerTypeAsync
}

// IsSync 检查是否为同步生产者
func (w *SaramaProducerWrapper) IsSync() bool {
	return w.typ == ProducerTypeSync
}
