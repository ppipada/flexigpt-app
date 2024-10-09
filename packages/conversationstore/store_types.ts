import { SecureSchema } from 'securejsondb';
import { Conversation as ConversationBase } from './conversation_types';

export type Conversation = ConversationBase & SecureSchema;
