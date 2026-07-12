import { useState } from "react";
import { act, renderHook } from "@testing-library/react";
import { useCursorStack } from "./useCursorStack";

function useTestHarness(initialCursor?: string) {
  const [cursor, setCursor] = useState<string | undefined>(initialCursor);
  const stack = useCursorStack(cursor, setCursor);
  return { cursor, ...stack };
}

describe("useCursorStack", () => {
  it("goNext pushes the pre-navigation cursor and advances", () => {
    const { result } = renderHook(() => useTestHarness());

    act(() => result.current.goNext("page2"));
    expect(result.current.cursor).toBe("page2");

    act(() => result.current.goNext("page3"));
    expect(result.current.cursor).toBe("page3");
  });

  it("goPrev pops and falls back to undefined on an empty stack", () => {
    const { result } = renderHook(() => useTestHarness());

    act(() => result.current.goPrev());
    expect(result.current.cursor).toBeUndefined();

    act(() => result.current.goNext("page2"));
    act(() => result.current.goNext("page3"));

    act(() => result.current.goPrev());
    expect(result.current.cursor).toBe("page2");

    act(() => result.current.goPrev());
    expect(result.current.cursor).toBeUndefined();
  });

  it("canGoPrev is true whenever an active cursor exists, even with an empty stack (deep-link case)", () => {
    const { result } = renderHook(() => useTestHarness("abc123"));
    expect(result.current.canGoPrev).toBe(true);
  });

  it("canGoPrev is false with no active cursor and an empty stack", () => {
    const { result } = renderHook(() => useTestHarness());
    expect(result.current.canGoPrev).toBe(false);
  });

  it("canGoPrev is true once the stack is non-empty", () => {
    const { result } = renderHook(() => useTestHarness());
    act(() => result.current.goNext("page2"));
    expect(result.current.canGoPrev).toBe(true);
  });

  it("reset clears the stack so a subsequent goPrev falls back to undefined", () => {
    const { result } = renderHook(() => useTestHarness());
    act(() => result.current.goNext("page2"));
    act(() => result.current.goNext("page3"));

    act(() => result.current.reset());
    act(() => result.current.goPrev());

    expect(result.current.cursor).toBeUndefined();
  });
});
