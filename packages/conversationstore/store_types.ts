import { Conversation as ConversationBase } from 'conversationmodel';
import { SecureSchema } from 'securejsondb';

export type Conversation = ConversationBase & SecureSchema;
