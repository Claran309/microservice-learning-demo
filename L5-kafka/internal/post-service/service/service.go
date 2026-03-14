package service

import (
	"context"
	"fmt"
	"log"
	"microservicesDemo/L5-kafka/internal/post-service/dao"
	"microservicesDemo/L5-kafka/internal/post-service/model"
	mq "microservicesDemo/L5-kafka/pkg/mq/kafka"
	"time"

	"github.com/cloudwego/hertz/pkg/common/json"
)

type PostService interface {
	//业务方法
	CreatePost(ctx context.Context, title string, userID int64, content string) (*model.Post, error)
	DeletePost(ctx context.Context, postID int64) error

	// 消费者方法
	StartConsumer(topic, groupID string) error
	StopConsumer()
}

type postServiceImpl struct {
	PostRepo      dao.PostRepository
	KafkaProducer *mq.Producer
	KafkaConsumer *mq.Consumer
	StopChan      chan struct{}
}

func NewPostServiceImpl(PostRepo dao.PostRepository, kafkaProducer *mq.Producer, kafkaConsumer *mq.Consumer) PostService {
	return &postServiceImpl{PostRepo: PostRepo, KafkaProducer: kafkaProducer, KafkaConsumer: kafkaConsumer, StopChan: make(chan struct{})}
}

// 启动消费者
func (s *postServiceImpl) StartConsumer(topic, groupID string) error {
	// 创建消费者
	consumer, err := mq.NewConsumer(topic, groupID)
	if err != nil {
		return fmt.Errorf("创建消费者失败: %v", err)
	}

	s.KafkaConsumer = consumer

	// 启动goroutine消费
	go s.consumeLoop()

	log.Printf("消费者启动: topic=%s, group=%s", topic, groupID)
	return nil
}

// 消费者主循环
func (s *postServiceImpl) consumeLoop() {
	for {
		select {
		case <-s.StopChan:
			log.Println("消费者停止")
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			// 接收消息
			msg, err := s.KafkaConsumer.Receive(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					// 读取超时，继续循环
					continue
				}
				log.Printf("接收消息失败: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if msg == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 处理消息
			go s.processMessage(msg)
		}
	}
}

// 处理消息
func (s *postServiceImpl) processMessage(msg *mq.Message) {
	// 解析消息
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Value, &data); err != nil {
		log.Printf("解析消息失败: %v", err)
		return
	}

	// 获取事件类型
	eventType, ok := data["event_type"].(string)
	if !ok {
		log.Printf("消息格式错误: 缺少event_type")
		return
	}

	// 根据事件类型调用已有的Service方法
	ctx := context.Background()

	switch eventType {
	case "USER_REGISTER":
		s.CreatePost(ctx, data["Welcome!"].(string), data["user_id"].(int64), data["content"].(string))
	case "USER_LOGIN":
		s.CreatePost(ctx, data["Login!"].(string), data["user_id"].(int64), data["content"].(string))
	default:
		log.Printf("未知事件类型: %s", eventType)
	}
}

// 停止消费者
func (s *postServiceImpl) StopConsumer() {
	if s.KafkaConsumer != nil {
		close(s.StopChan)
		s.KafkaConsumer.Close()
	}
}

func (s *postServiceImpl) CreatePost(ctx context.Context, title string, userID int64, content string) (*model.Post, error) {
	var post = model.Post{
		Title:   title,
		Content: content,
		Owner:   userID,
	}

	err := s.PostRepo.AddPost(ctx, &post)
	if err != nil {
		return nil, err
	}

	// 生产者发送消息
	if s.KafkaProducer != nil {
		eventData, _ := json.Marshal(map[string]interface{}{
			"event_type":    "POST_CREATED", // 明确的事件类型
			"post_id":       post.PostID,
			"user_id":       userID,
			"title":         post.Title,
			"content":       post.Content,
			"registered_at": time.Now(),
		})
		s.KafkaProducer.SendPostEvent(ctx, fmt.Sprintf("%d", post.PostID), eventData)
		fmt.Printf("发送帖子创建事件:user_id:%d", post.PostID)
	}

	return &post, nil
}

func (s *postServiceImpl) DeletePost(ctx context.Context, postID int64) error {
	return s.PostRepo.DeletePost(ctx, postID)
}
