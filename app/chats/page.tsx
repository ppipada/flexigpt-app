"use client";
import React, { useEffect, useRef, useState } from "react";
import { v7 as uuidv7 } from "uuid";
import ButtonScrollToBottom from "../components/ButtonScrollToBottom";
import { Chat, ChatCompletionRoleEnum, Message } from "../lib/models/ChatTypes";
import ChatInputField from "./ChatInputField";
import ChatMessage from "./ChatMessage";
import ChatNavBar from "./ChatNavbar";

const tmpDetails = `
start of details

\`\`\`dart
class ChatScreen extends StatefulWidget {
  const ChatScreen({super.key});

  @override
  ChatScreenState createState() => ChatScreenState();
}
\`\`\`

there is a \`console\`

\`\`\`py
def myFunc() {
  return "me";
}
\`\`\`
`;

const ChatScreen: React.FC = () => {
  const [chat, setChat] = useState<Chat>({
    id: uuidv7(),
    title: "New Chat",
    createTime: new Date(),
    modifiedTime: new Date(),
    messages: [],
  });

  const [inputHeight, setInputHeight] = useState(0);
  const chatContainerRef = useRef<HTMLDivElement>(null);

  const sendMessage = (messageContent: string) => {
    const trimmedText = messageContent.trim();
    if (trimmedText) {
      const newMessage: Message = {
        id: new Date().toISOString(),
        role: ChatCompletionRoleEnum.user,
        content: trimmedText,
        details: tmpDetails,
      };

      setChat((prevChat) => ({
        ...prevChat,
        messages: [...prevChat.messages, newMessage],
        modifiedTime: new Date(),
      }));
    }
  };

  useEffect(() => {
    if (chatContainerRef.current) {
      chatContainerRef.current.scrollTop =
        chatContainerRef.current.scrollHeight;
    }
  }, [chat.messages]);

  return (
    <div className="flex flex-col items-center w-full h-full overflow-hidden">
      <div className="w-full flex justify-center bg-transparent fixed top-0 z-10">
        <div className="w-11/12">
          <ChatNavBar
            title={chat.title}
            onNewChat={() => {
              setChat((prevChat) => ({
                ...prevChat,
                messages: [],
                modifiedTime: new Date(),
              }));
            }}
            onExport={() => {
              // Add export functionality
            }}
          />
        </div>
      </div>
      <div className="flex flex-col items-center w-full flex-grow bg-transparent overflow-hidden mt-16">
        <div
          className="w-full flex-grow flex justify-center overflow-y-auto"
          ref={chatContainerRef}
          style={{ maxHeight: `calc(100vh - 128px - ${inputHeight}px)` }} // Adjust height dynamically
        >
          <div className="w-11/12 lg:w-2/3">
            <div className="w-full flex-1 space-y-4">
              {chat.messages.map((message) => (
                <ChatMessage
                  key={message.id}
                  user={{
                    id: "1",
                    role: message.role,
                    icon: (
                      <div className="w-8 h-8 flex items-center justify-center bg-base-100 rounded-full">
                        {message.role === ChatCompletionRoleEnum.user
                          ? "U"
                          : "A"}
                      </div>
                    ),
                  }}
                  message={message}
                  onEdit={() => {}}
                  onResend={() => {}}
                  onLike={() => {}}
                  onDislike={() => {}}
                  onSendFeedback={() => {}}
                  feedbackController={React.createRef<HTMLInputElement>()}
                />
              ))}
            </div>
          </div>
        </div>
        <div className="w-full flex justify-center bg-transparent fixed bottom-0 z-10 mb-4">
          <div className="w-11/12 lg:w-2/3">
            <ChatInputField
              onSend={sendMessage}
              setInputHeight={setInputHeight}
            />
          </div>
        </div>
      </div>
      <div className="fixed bottom-0 right-0 mb-16 mr-16 z-10">
        <ButtonScrollToBottom
          scrollContainerRef={chatContainerRef}
          size={32}
          className="btn btn-md bg-transparent border-none flex items-center shadow-none"
        />
      </div>
    </div>
  );
};

export default ChatScreen;
