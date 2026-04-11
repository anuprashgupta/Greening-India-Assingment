import { Loader2 } from 'lucide-react';

interface Props {
  size?: number;
  className?: string;
  text?: string;
}

export default function LoadingSpinner({ size = 32, className = '', text }: Props) {
  return (
    <div className={`flex flex-col items-center justify-center gap-3 ${className}`}>
      <Loader2 size={size} className="animate-spin text-blue-500" />
      {text && <p className="text-sm text-gray-500 dark:text-gray-400">{text}</p>}
    </div>
  );
}
