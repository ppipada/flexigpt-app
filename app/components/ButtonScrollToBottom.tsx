import { useAtBottom } from "@/lib/hooks/use-at-bottom";
import React, { FC } from "react";
import { FiArrowDownCircle } from "react-icons/fi";

interface ButtonScrollToBottomProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  scrollContainerRef: React.RefObject<HTMLElement>;
  size: number;
}

const ButtonScrollToBottom: FC<ButtonScrollToBottomProps> = ({
  scrollContainerRef,
  size,
  ...props
}) => {
  const isAtBottom = useAtBottom(scrollContainerRef);

  return (
    !isAtBottom && (
      <button
        aria-label="Scroll to bottom"
        disabled={isAtBottom}
        onClick={() => {
          if (scrollContainerRef.current) {
            scrollContainerRef.current.scrollTo({
              top: scrollContainerRef.current.scrollHeight,
              behavior: "smooth",
            });
          }
        }}
        {...props}
      >
        <FiArrowDownCircle size={size} />
      </button>
    )
  );
};

export default ButtonScrollToBottom;
