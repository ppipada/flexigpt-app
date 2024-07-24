import * as React from "react";

export function useAtBottom(ref: React.RefObject<HTMLElement>, offset = 0) {
  const [isAtBottom, setIsAtBottom] = React.useState(false);

  React.useEffect(() => {
    const handleScroll = () => {
      if (ref.current) {
        const { scrollTop, scrollHeight, clientHeight } = ref.current;
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

  return isAtBottom;
}
