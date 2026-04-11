import { FolderOpen } from 'lucide-react';
import type { ReactNode } from 'react';

interface Props {
  icon?: ReactNode;
  title: string;
  description: string;
  action?: ReactNode;
}

export default function EmptyState({ icon, title, description, action }: Props) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-4 text-center">
      <div className="mb-4 text-gray-300 dark:text-gray-600">
        {icon || <FolderOpen size={64} />}
      </div>
      <h3 className="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-1">{title}</h3>
      <p className="text-sm text-gray-500 dark:text-gray-400 mb-6 max-w-sm">{description}</p>
      {action}
    </div>
  );
}
