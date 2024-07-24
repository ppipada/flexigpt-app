import React from "react";
import { ChatCompletionRoleEnum, Message, User } from "../lib/models/ChatTypes";
import { ChatMessageContent } from "./ChatMessageContent"; // Adjust the import path as necessary
import ChatMessageFooterArea from "./ChatMessageFooter"; // Adjust the import path as necessary

interface ChatMessageProps {
  user: User;
  message: Message;
  onEdit: () => void;
  onResend: () => void;
  onLike: () => void;
  onDislike: () => void;
  onSendFeedback: () => void;
  feedbackController: React.RefObject<HTMLInputElement>;
}

const ChatMessage: React.FC<ChatMessageProps> = ({
  user,
  message,
  onEdit,
  onResend,
  onLike,
  onDislike,
  onSendFeedback,
  feedbackController,
}) => {
  const isUser = user.role === ChatCompletionRoleEnum.user;
  const align = isUser ? "items-end text-right" : "items-start text-left";

  return (
    <div className={`flex-1 mb-4`}>
      <div className={`flex`}>
        {!isUser && (
          <div className="mt-0.5 mr-2">
            <div className="w-10 h-10">{user.icon}</div>
          </div>
        )}
        <div className="flex-1">
          <ChatMessageContent content={message.content} align={align} />
          <ChatMessageFooterArea
            isUser={isUser}
            onEdit={onEdit}
            onResend={onResend}
            onLike={onLike}
            onDislike={onDislike}
            onSendFeedback={onSendFeedback}
            feedbackController={feedbackController}
            messageDetails={message.details ? message.details : ""}
            cardContent={message.content}
          />
        </div>
        {isUser && (
          <div className="mt-0.5 ml-2">
            <div className="w-10 h-10">{user.icon}</div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ChatMessage;
