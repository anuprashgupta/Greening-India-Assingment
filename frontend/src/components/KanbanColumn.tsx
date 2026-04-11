import { useDroppable } from '@dnd-kit/core';
import type { ReactNode } from 'react';

interface Props {
  id: string;
  title: string;
  color: string;
  count: number;
  children: ReactNode;
}

export default function KanbanColumn({ id, title, color, count, children }: Props) {
  const { isOver, setNodeRef } = useDroppable({ id });

  return (
    <div
      ref={setNodeRef}
      className={`flex flex-col rounded-xl border-t-4 ${color} bg-gray-50 dark:bg-gray-800/50 p-3 min-h-[200px] transition-colors ${
        isOver ? 'bg-blue-50 dark:bg-blue-900/20 ring-2 ring-blue-300 dark:ring-blue-600' : ''
      }`}
    >
      <div className="mb-3 flex items-center justify-between px-1">
        <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">{title}</h3>
        <span className="rounded-full bg-gray-200 dark:bg-gray-700 px-2 py-0.5 text-xs font-medium text-gray-600 dark:text-gray-400">
          {count}
        </span>
      </div>
      <div className="flex flex-1 flex-col gap-2">
        {children}
      </div>
    </div>
  );
}
