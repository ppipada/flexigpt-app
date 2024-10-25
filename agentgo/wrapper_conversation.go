package main

import (
	"context"

	"github.com/flexigpt/flexiui/pkggo/conversationstore"
	"github.com/flexigpt/flexiui/pkggo/conversationstore/spec"
)

type ConversationCollectionWrapper struct {
	store *conversationstore.ConversationCollection
}

func InitConversationCollectionWrapper(
	c *ConversationCollectionWrapper,
	conversationDir string,
) error {
	conversationStoreAPI := &conversationstore.ConversationCollection{}
	err := conversationstore.InitConversationCollection(conversationStoreAPI, conversationDir)
	if err != nil {
		return err
	}
	c.store = conversationStoreAPI
	return nil
}

func (ccw *ConversationCollectionWrapper) SaveConversation(
	req *spec.SaveConversationRequest,
) (*spec.SaveConversationResponse, error) {
	return ccw.store.SaveConversation(context.Background(), req)
}

func (ccw *ConversationCollectionWrapper) DeleteConversation(
	req *spec.DeleteConversationRequest,
) (*spec.DeleteConversationResponse, error) {

	return ccw.store.DeleteConversation(context.Background(), req)
}

func (ccw *ConversationCollectionWrapper) GetConversation(
	req *spec.GetConversationRequest,
) (*spec.GetConversationResponse, error) {

	return ccw.store.GetConversation(context.Background(), req)
}

func (ccw *ConversationCollectionWrapper) ListConversations(
	req *spec.ListConversationsRequest,
) (*spec.ListConversationsResponse, error) {

	return ccw.store.ListConversations(context.Background(), req)
}

func (ccw *ConversationCollectionWrapper) AddMessageToConversation(
	req *spec.AddMessageToConversationRequest,
) (*spec.AddMessageToConversationResponse, error) {

	return ccw.store.AddMessageToConversation(context.Background(), req)
}
