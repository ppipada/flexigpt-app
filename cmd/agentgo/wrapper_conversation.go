package main

import (
	"context"

	"github.com/ppipada/flexigpt-app/pkg/conversationstore"
	"github.com/ppipada/flexigpt-app/pkg/conversationstore/spec"
	"github.com/ppipada/flexigpt-app/pkg/middleware"
)

type ConversationCollectionWrapper struct {
	store *conversationstore.ConversationCollection
}

func InitConversationCollectionWrapper(
	c *ConversationCollectionWrapper,
	conversationDir string,
) error {
	conversationStoreAPI, err := conversationstore.NewConversationCollection(
		conversationDir,
		conversationstore.WithFTS(true),
	)
	if err != nil {
		return err
	}
	c.store = conversationStoreAPI
	return nil
}

func (ccw *ConversationCollectionWrapper) PutConversation(
	req *spec.SaveConversationRequest,
) (*spec.SaveConversationResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.SaveConversationResponse, error) {
		return ccw.store.PutConversation(context.Background(), req)
	})
}

func (ccw *ConversationCollectionWrapper) DeleteConversation(
	req *spec.DeleteConversationRequest,
) (*spec.DeleteConversationResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.DeleteConversationResponse, error) {
		return ccw.store.DeleteConversation(context.Background(), req)
	})
}

func (ccw *ConversationCollectionWrapper) GetConversation(
	req *spec.GetConversationRequest,
) (*spec.GetConversationResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.GetConversationResponse, error) {
		return ccw.store.GetConversation(context.Background(), req)
	})
}

func (ccw *ConversationCollectionWrapper) ListConversations(
	req *spec.ListConversationsRequest,
) (*spec.ListConversationsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.ListConversationsResponse, error) {
		return ccw.store.ListConversations(context.Background(), req)
	})
}

func (ccw *ConversationCollectionWrapper) SearchConversations(
	req *spec.SearchConversationsRequest,
) (*spec.SearchConversationsResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.SearchConversationsResponse, error) {
		return ccw.store.SearchConversations(context.Background(), req)
	})
}

func (ccw *ConversationCollectionWrapper) PutMessagesToConversation(
	req *spec.PutMessagesToConversationRequest,
) (*spec.PutMessagesToConversationResponse, error) {
	return middleware.WithRecoveryResp(func() (*spec.PutMessagesToConversationResponse, error) {
		return ccw.store.PutMessagesToConversation(context.Background(), req)
	})
}
