import { AlertIcon } from "./ui/icons";

interface ErrorBannerProps {
  message: string;
  onRetry?: () => void;
}

export function ErrorBanner({ message, onRetry }: ErrorBannerProps) {
  return (
    <div
      role="alert"
      className="flex items-center justify-between gap-4 rounded-xl border border-red-200 bg-red-50 px-4 py-3.5 text-sm text-red-800"
    >
      <span className="flex items-center gap-2.5">
        <AlertIcon className="h-5 w-5 shrink-0 text-red-500" />
        {message}
      </span>
      {onRetry && (
        <button
          type="button"
          onClick={onRetry}
          className="shrink-0 rounded-lg border border-red-200 bg-white px-3 py-1.5 font-medium text-red-700 shadow-sm transition hover:bg-red-100"
        >
          Retry
        </button>
      )}
    </div>
  );
}
