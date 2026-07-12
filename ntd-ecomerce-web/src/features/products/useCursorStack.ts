import { useRef } from "react";

export interface CursorStack {
  canGoPrev: boolean;
  goNext: (nextCursor: string) => void;
  goPrev: () => void;
  reset: () => void;
}

export function useCursorStack(
  cursor: string | undefined,
  setCursor: (cursor: string | undefined) => void,
): CursorStack {
  const stackRef = useRef<Array<string | undefined>>([]);

  function goNext(nextCursor: string) {
    stackRef.current.push(cursor);
    setCursor(nextCursor);
  }

  function goPrev() {
    const previous = stackRef.current.pop();
    setCursor(previous);
  }

  function reset() {
    stackRef.current = [];
  }

  return {
    canGoPrev: stackRef.current.length > 0 || cursor !== undefined,
    goNext,
    goPrev,
    reset,
  };
}
