import React, { FC } from "react";
import { FiArrowDownCircle } from "react-icons/fi";

function useAtBottom(ref: React.RefObject<HTMLElement>, offset = 0) {
  const [isAtBottom, setIsAtBottom] = React.useState(false);
  const [isScrollable, setIsScrollable] = React.useState(false);

  React.useEffect(() => {
    const handleScroll = () => {
      if (ref.current) {
        const { scrollTop, scrollHeight, clientHeight } = ref.current;
        setIsScrollable(scrollHeight > clientHeight);
        setIsAtBottom(scrollTop + clientHeight >= scrollHeight - offset - 10);
      }
    };

    // Add a throttling mechanism
    let throttleTimeout: NodeJS.Timeout | null = null;
    const throttledHandleScroll = () => {
      if (throttleTimeout === null) {
        throttleTimeout = setTimeout(() => {
          handleScroll();
          throttleTimeout = null;
        }, 100); // Adjust the timeout value as needed
      }
    };

    if (ref.current) {
      ref.current.addEventListener("scroll", throttledHandleScroll, {
        passive: true,
      });
    }
    handleScroll();

    return () => {
      if (ref.current) {
        ref.current.removeEventListener("scroll", throttledHandleScroll);
      }
      if (throttleTimeout) {
        clearTimeout(throttleTimeout);
      }
    };
  }, [ref, offset]);

  return { isAtBottom, isScrollable };
}

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
  const { isAtBottom, isScrollable } = useAtBottom(scrollContainerRef);

  return (
    isScrollable &&
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
