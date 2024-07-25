import { ChatMessageContent } from "@/chats/ChatMessageContent"; // Adjust the import path as necessary
import ChatMessageFooterArea from "@/chats/ChatMessageFooter"; // Adjust the import path as necessary
import { ChatCompletionRoleEnum, Message, User } from "@/lib/models/ChatTypes";
import React from "react";

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
  const align = !isUser ? "items-end text-left" : "items-start text-left";
  const leftColSpan = !isUser ? "col-span-1 lg:col-span-2" : "col-span-1";
  const rightColSpan = !isUser ? "col-span-1" : "col-span-1 lg:col-span-2";

  return (
    <div className="grid grid-cols-12 gap-2 mb-4 items-start">
      <div className={`${leftColSpan} flex justify-end`}>
        {isUser && <div className="w-10 h-10 mt-0.5 pl-1">{user.icon}</div>}
      </div>
      <div className="col-span-10 lg:col-span-9">
        <div className="flex flex-col w-full">
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
      </div>
      <div className={`${rightColSpan} flex justify-start`}>
        {!isUser && <div className="w-10 h-10 mt-0.5 pl-1">{user.icon}</div>}
      </div>
    </div>
  );
};

export default ChatMessage;
