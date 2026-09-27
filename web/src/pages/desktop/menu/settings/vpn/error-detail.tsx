type ErrorDetailProps = {
  message: string;
};

// ErrorDetail shows why an action failed. The server sends the last lines of
// the command's output, so the line breaks are kept.
export const ErrorDetail = ({ message }: ErrorDetailProps) => {
  if (!message) return null;

  return (
    <pre className="mt-5 max-h-[200px] w-full overflow-auto rounded bg-neutral-800/60 p-3 font-mono text-xs break-words whitespace-pre-wrap text-red-400">
      {message}
    </pre>
  );
};
